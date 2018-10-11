package cmd

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	helm_env "k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/hooks"

	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/types"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/status"

	"github.com/ContainerSolutions/helm-convert/pkg/generators"
	"github.com/ContainerSolutions/helm-convert/pkg/helm"
	"github.com/ContainerSolutions/helm-convert/pkg/transformers"
)

var (
	whitespaceRegex = regexp.MustCompile(`^\s*$`)
	settings        helm_env.EnvSettings
)

const defaultDirectoryPermission = 0755

type convertCmd struct {
	home helmpath.Home

	chart        string
	repoURL      string
	destination  string
	name         string
	namespace    string
	fileValues   []string
	valueFiles   helm.ValueFiles
	values       []string
	stringValues []string
	version      string
	depUp        bool

	username string
	password string
	certFile string
	keyFile  string
	caFile   string

	verify      bool
	verifyLater bool
	keyring     string

	out io.Writer
}

const convertDesc = `
This command convert a Helm chart into a kustomize compatible package.
`

const convertExample = `
  # convert the stable/mongodb chart
  helm convert --destination mongodb --name mongodb stable/mongodb

  # convert chart from a url
  helm convert https://s3-eu-west-1.amazonaws.com/coreos-charts/stable/prometheus-operator

  # convert the stable/mongodb chart with a given values.yaml file
  helm convert -f values.yaml stable/mongodb

  # convert the stable/mongodb chart and override values using --set flag:
  helm convert --set persistence.enabled=true stable/mongodb
`

func NewConvertCommand() *cobra.Command {
	k := &convertCmd{
		out: os.Stdout,
	}

	c := &cobra.Command{
		Use:     "convert [flag] [chart URL | repo/chartname] [...]",
		Short:   "convert a chart",
		Long:    convertDesc,
		Example: convertExample,
		Args:    cobra.MinimumNArgs(1),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			flag.CommandLine.Parse([]string{})
		},
		RunE: func(c *cobra.Command, args []string) error {
			settings.Home = k.home

			for i := 0; i < len(args); i++ {
				k.chart = args[i]
				if err := k.run(); err != nil {
					return err
				}
			}

			return nil
		},
	}

	f := c.Flags()
	f.StringVar(&k.name, "name", "", "release name")
	f.VarP(&k.valueFiles, "values", "f", "specify values in a YAML file or a URL(can specify multiple)")
	f.StringVar((*string)(&k.home), "home", helm_env.DefaultHelmHome, "location of your Helm config. Overrides $HELM_HOME")
	f.StringArrayVar(&k.values, "set", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringArrayVar(&k.fileValues, "set-file", []string{}, "set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)")
	f.StringArrayVar(&k.stringValues, "set-string", []string{}, "set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.BoolVar(&k.verify, "verify", false, "verify the package against its signature")
	f.BoolVar(&k.verifyLater, "prov", false, "fetch the provenance file, but don't perform verification")
	f.StringVar(&k.version, "version", "", "specific version of a chart. Without this, the latest version is fetched")
	f.StringVar(&k.keyring, "keyring", defaultKeyring(), "keyring containing public keys")
	f.StringVarP(&k.destination, "destination", "d", "", "location to write the chart. If this and tardir are specified, tardir is appended to this")
	f.StringVar(&k.repoURL, "repo", "", "chart repository url where to locate the requested chart")
	f.StringVar(&k.certFile, "cert-file", "", "identify HTTPS client using this SSL certificate file")
	f.StringVar(&k.keyFile, "key-file", "", "identify HTTPS client using this SSL key file")
	f.StringVar(&k.caFile, "ca-file", "", "verify certificates of HTTPS-enabled servers using this CA bundle")
	f.BoolVar(&k.depUp, "dep-up", false, "run helm dependency update before installing the chart")
	f.StringVar(&k.username, "username", "", "chart repository username")
	f.StringVar(&k.password, "password", "", "chart repository password")

	// log to stderr by default,
	flag.Set("logtostderr", "true")

	// add glog flags
	c.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	return c
}

func (k *convertCmd) run() error {
	h := helm.NewHelm(settings, k.out)

	glog.V(8).Infof("Using settings %#v", settings)

	// load chart
	chartRequested, err := h.LoadChart(&helm.LoadChartConfig{
		RepoURL:  k.repoURL,
		Username: k.username,
		Password: k.password,
		Chart:    k.chart,
		Version:  k.version,
		DepUp:    k.depUp,
		Verify:   k.verify,
		Keyring:  k.keyring,
		CertFile: k.certFile,
		KeyFile:  k.keyFile,
		CaFile:   k.caFile,
	})
	if err != nil {
		return prettyError(err)
	}

	// use chart name if destination isn't defined via flags
	if k.destination == "" {
		k.destination = chartRequested.Metadata.Name
	}

	// use chart name if name isn't defined via flags
	if k.name == "" {
		k.name = chartRequested.Metadata.Name
	}

	// render charts with given values
	renderedManifests, err := h.RenderChart(&helm.RenderChartConfig{
		ChartRequested: chartRequested,
		Name:           k.name,
		Namespace:      k.namespace,
		ValueFiles:     k.valueFiles,
		Values:         k.values,
		StringValues:   k.stringValues,
		FileValues:     k.fileValues,
	})
	if err != nil {
		return prettyError(err)
	}

	// convert Yaml to resource
	resources := resmap.ResMap{}
	for _, m := range renderedManifests {
		data := m.Content
		b := filepath.Base(m.Name)
		if b == "NOTES.txt" {
			continue
		}
		if whitespaceRegex.MatchString(data) {
			continue
		}
		if strings.HasPrefix(b, "_") {
			continue
		}

		r, err := newResource([]byte(data))
		if err != nil {
			glog.Fatalf("Error converting yaml to resources: %v", err)
		}
		resources[r.Id()] = r
	}

	config := &types.Kustomization{}

	// initialize transformers
	r := []transformers.Transformer{
		transformers.NewLabelsTransformer([]string{"chart", "release"}),
		transformers.NewAnnotationsTransformer([]string{
			hooks.HookAnno,
			hooks.HookWeightAnno,
			hooks.HookDeleteAnno,
		}),
		transformers.NewImageTagTransformer(),
		transformers.NewSecretTransformer(),
		transformers.NewNamePrefixTransformer(),
		transformers.NewResourcesTransformer(),
		transformers.NewEmptyTransformer(),
	}

	// gather kustomization config via transformers
	err = transformers.NewMultiTransformer(r).Transform(config, resources)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// write to disk
	generator := generators.NewGenerator()
	err = generator.Render(k.destination, config, chartRequested.Metadata, resources)
	if err != nil {
		return err
	}

	return nil
}

func newResource(in []byte) (output *resource.Resource, err error) {
	m := map[string]interface{}{}

	err = yaml.Unmarshal(in, &m)
	if err != nil {
		return
	}

	output = resource.NewResourceFromMap(m)
	return
}

func prettyError(err error) error {
	if err == nil {
		return nil
	}
	if s, ok := status.FromError(err); ok {
		return fmt.Errorf(s.Message())
	}
	return err
}

// defaultKeyring returns the expanded path to the default keyring.
func defaultKeyring() string {
	return os.ExpandEnv("$HOME/.gnupg/pubring.gpg")
}

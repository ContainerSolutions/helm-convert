package cmd

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ContainerSolutions/helm-convert/pkg/generators"
	"github.com/ContainerSolutions/helm-convert/pkg/helm"
	"github.com/ContainerSolutions/helm-convert/pkg/transformers"
	"github.com/ContainerSolutions/helm-convert/pkg/types"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/status"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	helm_env "k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/hooks"
	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/pkg/resource"
	ktypes "sigs.k8s.io/kustomize/pkg/types"
)

var (
	whitespaceRegex = regexp.MustCompile(`^\s*$`)
	settings        helm_env.EnvSettings
)

const defaultDirectoryPermission = 0755

type convertCmd struct {
	home helmpath.Home

	chart               string
	repoURL             string
	destination         string
	resourceDestination string
	name                string
	namespace           string
	fileValues          []string
	valueFiles          helm.ValueFiles
	values              []string
	stringValues        []string
	skipTransformers    []string
	version             string
	depUp               bool
	forceGen            bool
	comments            bool

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

// NewConvertCommand constructs a new convert command
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
	f.StringSliceVar(&k.skipTransformers, "skip-transformers", []string{}, "set a list of transformers that are skipped during the conversion process (can specify multiple or separate values with commas: secret,configmap)")
	f.BoolVar(&k.verify, "verify", false, "verify the package against its signature")
	f.BoolVar(&k.verifyLater, "prov", false, "fetch the provenance file, but don't perform verification")
	f.StringVar(&k.namespace, "namespace", "default", "global namespace to use for the manifests")
	f.StringVar(&k.version, "version", "", "specific version of a chart. Without this, the latest version is fetched")
	f.StringVar(&k.keyring, "keyring", defaultKeyring(), "keyring containing public keys")
	f.StringVarP(&k.destination, "destination", "d", "", "location to write the chart. If this and tardir are specified, tardir is appended to this")
	f.StringVarP(&k.resourceDestination, "resource-destination", "r", "", "location to write the resources.")
	f.StringVar(&k.repoURL, "repo", "", "chart repository url where to locate the requested chart")
	f.StringVar(&k.certFile, "cert-file", "", "identify HTTPS client using this SSL certificate file")
	f.StringVar(&k.keyFile, "key-file", "", "identify HTTPS client using this SSL key file")
	f.StringVar(&k.caFile, "ca-file", "", "verify certificates of HTTPS-enabled servers using this CA bundle")
	f.BoolVar(&k.depUp, "dep-up", false, "run helm dependency update before installing the chart")
	f.BoolVar(&k.forceGen, "force", false, "convert chart even if the destination directory already exists")
	f.StringVar(&k.username, "username", "", "chart repository username")
	f.StringVar(&k.password, "password", "", "chart repository password")
	f.BoolVar(&k.comments, "comments", true, "add default comments to kustomization.yaml file")

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

	if k.resourceDestination != "" {
		os.MkdirAll(k.resourceDestination, 0755)
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
	resources := types.NewResources()
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

		resList, err := newResources([]byte(data))
		if err != nil {
			glog.Fatalf("Error converting yaml to resources: %v", err)
		}
		for _, r := range resList {
			// if k.resourceDestination != "" {
			// 	r.SetName(path.Join(k.resourceDestination, r.GetName()))
			// }
			resources.ResMap[r.Id()] = r
		}
	}

	config := &ktypes.Kustomization{}

	defaultTransfomers := []transformers.Transformer{
		transformers.NewLabelsTransformer([]string{"chart", "release", "heritage"}),
		transformers.NewAnnotationsTransformer([]string{
			hooks.HookAnno,
			hooks.HookWeightAnno,
			hooks.HookDeleteAnno,
		}),
		transformers.NewImageTransformer(),
		transformers.NewConfigMapTransformer(),
		transformers.NewSecretTransformer(),
		transformers.NewNamePrefixTransformer(k.name),
		transformers.NewResourcesTransformer(k.resourceDestination),
		transformers.NewEmptyTransformer(),
	}

	// load transformers
	var r []transformers.Transformer
	if len(k.skipTransformers) > 0 {
		skipMap := make(map[string]struct{}, len(k.skipTransformers))
		for _, s := range k.skipTransformers {
			skipMap[strings.ToLower(s)] = struct{}{}
		}

		r = make([]transformers.Transformer, 0, len(defaultTransfomers))
		for _, dt := range defaultTransfomers {
			if _, ok := skipMap[transformerName(dt)]; !ok {
				r = append(r, dt)
			}
		}
	} else {
		r = defaultTransfomers
	}

	// gather kustomization config via transformers
	err = transformers.NewMultiTransformer(r).Transform(config, resources)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// write to disk
	generator := generators.NewGenerator(k.forceGen, k.resourceDestination)
	err = generator.Render(k.destination, config, chartRequested.Metadata, resources, k.comments)
	if err != nil {
		return err
	}

	return nil
}

func newResources(in []byte) ([]*resource.Resource, error) {
	decoder := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(in), 1024)
	rf := resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())

	var result []*resource.Resource
	var err error
	for err == nil || isEmptyYamlError(err) {
		var out map[string]interface{}
		err = decoder.Decode(&out)
		if err == nil {
			// ignore empty chunks
			if len(out) == 0 {
				continue
			}

			if list, ok := isList(out); ok {
				for _, i := range list {
					if item, ok := i.(map[string]interface{}); ok {
						result = append(result, rf.FromMap(item))
					}
				}
			} else {
				result = append(result, rf.FromMap(out))
			}
		}
	}
	if err != io.EOF {
		return nil, err
	}
	return result, nil
}

func isEmptyYamlError(err error) bool {
	return strings.Contains(err.Error(), "is missing in 'null'")
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

func isList(res map[string]interface{}) ([]interface{}, bool) {
	itemList, ok := res["items"]
	if !ok {
		return nil, false
	}

	items, ok := itemList.([]interface{})
	return items, ok
}

func transformerName(t transformers.Transformer) string {
	return strings.ToLower(
		strings.TrimSuffix(
			strings.TrimPrefix(
				fmt.Sprintf("%T", t),
				"*transformers."),
			"Transformer"))
}

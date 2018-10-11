// Package helm handle helm resources
package helm

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/downloader"
	"k8s.io/helm/pkg/getter"
	helm_env "k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/manifest"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/renderutil"
	"k8s.io/helm/pkg/repo"
	"k8s.io/helm/pkg/strvals"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"
)

var defaultKubeVersion = fmt.Sprintf("%s.%s", chartutil.DefaultKubeVersion.Major, chartutil.DefaultKubeVersion.Minor)

// Helm type
type Helm struct {
	settings helm_env.EnvSettings
	out      io.Writer
}

// LoadChartConfig define the configuration to load a chart
type LoadChartConfig struct {
	RepoURL  string
	Username string
	Password string
	Chart    string
	Version  string
	DepUp    bool
	Verify   bool
	Keyring  string
	CertFile string
	KeyFile  string
	CaFile   string
}

// RenderChartConfig define the configuration to render a chart
type RenderChartConfig struct {
	ChartRequested *chart.Chart
	Name           string
	Namespace      string
	ValueFiles     ValueFiles
	Values         []string
	StringValues   []string
	FileValues     []string
}

// NewHelm constructs helm
func NewHelm(settings helm_env.EnvSettings, out io.Writer) *Helm {
	return &Helm{
		settings,
		out,
	}
}

// LoadChart download a chart or load it from cache
func (h *Helm) LoadChart(c *LoadChartConfig) (*chart.Chart, error) {
	glog.V(8).Infof("Loading chart with settings %#v", c)

	chartPath, err := h.LocateChartPath(
		c.RepoURL,
		c.Username,
		c.Password,
		c.Chart,
		c.Version,
		c.Verify,
		c.Keyring,
		c.CertFile,
		c.KeyFile,
		c.CaFile,
	)
	if err != nil {
		return nil, err
	}
	glog.V(8).Infof("Using chart path %v", chartPath)

	// Check chart requirements to make sure all dependencies are present in /charts
	chartRequested, err := chartutil.Load(chartPath)
	if err != nil {
		return nil, err
	}

	if req, err := chartutil.LoadRequirements(chartRequested); err == nil {
		if err := renderutil.CheckDependencies(chartRequested, req); err != nil {
			if c.DepUp {
				man := &downloader.Manager{
					Out:        h.out,
					ChartPath:  chartPath,
					HelmHome:   h.settings.Home,
					Keyring:    c.Keyring,
					SkipUpdate: false,
					Getters:    getter.All(h.settings),
				}
				if err := man.Update(); err != nil {
					return nil, err
				}

				// Update all dependencies which are present in /charts.
				chartRequested, err = chartutil.Load(chartPath)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}
	} else if err != chartutil.ErrRequirementsNotFound {
		return nil, fmt.Errorf("cannot load requirements: %v", err)
	}

	return chartRequested, nil
}

// RenderChart manifest
func (h *Helm) RenderChart(c *RenderChartConfig) ([]manifest.Manifest, error) {
	renderOpts := renderutil.Options{
		ReleaseOptions: chartutil.ReleaseOptions{
			Name:      c.Name,
			Namespace: c.Namespace,
		},
		KubeVersion: defaultKubeVersion,
	}
	glog.V(8).Infof("Rendering chart with options: %#v\n", renderOpts)

	// get combined values and create config
	rawVals, err := h.Vals(c.ValueFiles, c.Values, c.StringValues, c.FileValues, "", "", "")
	if err != nil {
		return nil, err
	}

	config := &chart.Config{Raw: string(rawVals), Values: map[string]*chart.Value{}}
	glog.V(10).Infof("Chart config: %#v", config)
	glog.V(10).Info("Chart requested", c.ChartRequested)

	renderedTemplates, err := renderutil.Render(c.ChartRequested, config, renderOpts)
	if err != nil {
		return nil, err
	}

	return manifest.SplitManifests(renderedTemplates), nil
}

// LocateChartPath looks for a chart directory in known places, and returns either the full path or an error.
//
// This does not ensure that the chart is well-formed; only that the requested filename exists.
//
// Order of resolution:
// - current working directory
// - if path is absolute or begins with '.', error out here
// - chart repos in $HELM_HOME
// - URL
//
// If 'verify' is true, this will attempt to also verify the chart.
func (h *Helm) LocateChartPath(repoURL, username, password, name, version string, verify bool, keyring,
	certFile, keyFile, caFile string) (string, error) {
	name = strings.TrimSpace(name)
	version = strings.TrimSpace(version)
	if fi, err := os.Stat(name); err == nil {
		abs, err := filepath.Abs(name)
		if err != nil {
			return abs, err
		}
		if verify {
			if fi.IsDir() {
				return "", errors.New("cannot verify a directory")
			}
			if _, err := downloader.VerifyChart(abs, keyring); err != nil {
				return "", err
			}
		}
		return abs, nil
	}

	if filepath.IsAbs(name) || strings.HasPrefix(name, ".") {
		return name, fmt.Errorf("path %q not found", name)
	}

	crepo := filepath.Join(h.settings.Home.Repository(), name)

	if _, err := os.Stat(crepo); err == nil {
		return filepath.Abs(crepo)
	}

	dl := downloader.ChartDownloader{
		HelmHome: h.settings.Home,
		Out:      h.out,
		Keyring:  keyring,
		Getters:  getter.All(h.settings),
		Username: username,
		Password: password,
	}

	if verify {
		dl.Verify = downloader.VerifyAlways
	}

	if repoURL != "" {
		chartURL, err := repo.FindChartInAuthRepoURL(repoURL, username, password, name, version,
			certFile, keyFile, caFile, getter.All(h.settings))
		if err != nil {
			return "", err
		}
		name = chartURL
	}

	if _, err := os.Stat(h.settings.Home.Archive()); os.IsNotExist(err) {
		os.MkdirAll(h.settings.Home.Archive(), 0744)
	}

	glog.V(8).Infof("Downloading chart '%s' version '%s' with parameters: %#v\n", name, version, dl)
	filename, _, err := dl.DownloadTo(name, version, h.settings.Home.Archive())

	if err != nil {
		return filename, err
	}

	lname, err := filepath.Abs(filename)
	if err != nil {
		return filename, err
	}
	glog.V(4).Infof("Fetched %s to %s\n", name, filename)

	return lname, nil
}

// Vals merges values from files specified via -f/--values and
// directly via --set or --set-string or --set-file, marshaling them to YAML
func (h *Helm) Vals(valueFiles ValueFiles, values []string, stringValues []string, fileValues []string, CertFile, KeyFile, CAFile string) ([]byte, error) {
	base := map[string]interface{}{}

	// User specified a values files via -f/--values
	for _, filePath := range valueFiles {
		currentMap := map[string]interface{}{}

		var bytes []byte
		var err error
		if strings.TrimSpace(filePath) == "-" {
			bytes, err = ioutil.ReadAll(os.Stdin)
		} else {
			bytes, err = h.readFile(filePath, CertFile, KeyFile, CAFile)
		}

		if err != nil {
			return []byte{}, err
		}

		if err := yaml.Unmarshal(bytes, &currentMap); err != nil {
			return []byte{}, fmt.Errorf("failed to parse %s: %s", filePath, err)
		}
		// Merge with the previous map
		base = mergeValues(base, currentMap)
	}

	// User specified a value via --set
	for _, value := range values {
		if err := strvals.ParseInto(value, base); err != nil {
			return []byte{}, fmt.Errorf("failed parsing --set data: %s", err)
		}
	}

	// User specified a value via --set-string
	for _, value := range stringValues {
		if err := strvals.ParseIntoString(value, base); err != nil {
			return []byte{}, fmt.Errorf("failed parsing --set-string data: %s", err)
		}
	}

	// User specified a value via --set-file
	for _, value := range fileValues {
		reader := func(rs []rune) (interface{}, error) {
			bytes, err := h.readFile(string(rs), CertFile, KeyFile, CAFile)
			return string(bytes), err
		}
		if err := strvals.ParseIntoFile(value, base, reader); err != nil {
			return []byte{}, fmt.Errorf("failed parsing --set-file data: %s", err)
		}
	}

	return yaml.Marshal(base)
}

//readFile load a file from the local directory or a remote file with a url.
func (h *Helm) readFile(filePath, CertFile, KeyFile, CAFile string) ([]byte, error) {
	u, _ := url.Parse(filePath)
	p := getter.All(h.settings)

	// FIXME: maybe someone handle other protocols like ftp.
	getterConstructor, err := p.ByScheme(u.Scheme)

	if err != nil {
		return ioutil.ReadFile(filePath)
	}

	getter, err := getterConstructor(filePath, CertFile, KeyFile, CAFile)
	if err != nil {
		return []byte{}, err
	}
	data, err := getter.Get(filePath)
	return data.Bytes(), err
}

// Merges source and destination map, preferring values from the source map
func mergeValues(dest map[string]interface{}, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = v
			continue
		}
		nextMap, ok := v.(map[string]interface{})
		// If it isn't another map, overwrite the value
		if !ok {
			dest[k] = v
			continue
		}
		// Edge case: If the key exists in the destination, but isn't a map
		destMap, isMap := dest[k].(map[string]interface{})
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest[k] = v
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		dest[k] = mergeValues(destMap, nextMap)
	}
	return dest
}

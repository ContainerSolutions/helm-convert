package transformers

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/ContainerSolutions/helm-convert/pkg/types"
	"github.com/golang/glog"
	ktypes "sigs.k8s.io/kustomize/pkg/types"
)

type configMapTransformer struct{}

var regexpMultiline = regexp.MustCompile("\n")

var _ Transformer = &configMapTransformer{}

// NewConfigMapTransformer constructs a configMapTransformer.
func NewConfigMapTransformer() Transformer {
	return &configMapTransformer{}
}

// Transform retrieve configmap from manifests and store them as configMapGenerator in the kustomization.yaml
func (t *configMapTransformer) Transform(config *ktypes.Kustomization, resources *types.Resources) error {
	for id, res := range resources.ResMap {
		kind, err := res.GetFieldValue("kind")
		if err != nil {
			return err
		}

		if kind != "ConfigMap" {
			continue
		}

		name, err := res.GetFieldValue("metadata.name")
		if err != nil {
			return err
		}

		obj := resources.ResMap[id].Map()

		if _, found := obj["data"]; !found {
			glog.V(8).Infof("Data field from configmap '%s' was not found", name)
			continue
		}

		if obj["data"] == nil {
			glog.V(8).Infof("Data field from configmap '%s' is empty", name)
			continue
		}

		data := obj["data"].(map[string]interface{})

		configMapArg := ktypes.ConfigMapArgs{
			GeneratorArgs: ktypes.GeneratorArgs{
				Name: name,
			},
		}

		var fileSources []string
		var literalSources []string

		for key, value := range data {
			// if multiline, store as external file otherwise literal
			// TODO: detect if key/value file, ie: .env, .ini and set DataSources.EnvSource
			if v, ok := value.(string); ok {
				if regexpMultiline.MatchString(v) {
					glog.V(8).Infof("Converting '%s' as external file from configmap '%s'", key, name)
					filename := fmt.Sprintf("%s-%s", name, key)
					fileSources = append(fileSources, filename)
					resources.ConfigFiles[filename] = v
				} else {
					glog.V(8).Infof("Converting '%s' as literal value from configmap '%s'", key, name)
					literalSources = append(literalSources, fmt.Sprintf("%s=\"%s\"", key, value))
				}
			}
		}

		sort.Strings(literalSources)
		sort.Strings(fileSources)

		configMapArg.GeneratorArgs.DataSources = ktypes.DataSources{
			LiteralSources: literalSources,
			FileSources:    fileSources,
		}

		config.ConfigMapGenerator = append(config.ConfigMapGenerator, configMapArg)
		delete(resources.ResMap, res.Id())
	}

	return nil
}

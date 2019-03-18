package transformers

import (
	"regexp"

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
		dataMap := make(map[string]string, len(data))
		for key, value := range data {
			dataMap[key] = value.(string)
		}

		configMapArg := ktypes.ConfigMapArgs{
			GeneratorArgs: ktypes.GeneratorArgs{
				Name: name,
			},
		}

		configMapArg.GeneratorArgs.DataSources = TransformDataSource(name, dataMap, resources.SourceFiles)

		config.ConfigMapGenerator = append(config.ConfigMapGenerator, configMapArg)
		delete(resources.ResMap, res.Id())
	}

	return nil
}

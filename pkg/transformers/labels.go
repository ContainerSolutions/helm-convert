package transformers

import (
	ktypes "sigs.k8s.io/kustomize/pkg/types"

	"github.com/ContainerSolutions/helm-convert/pkg/types"
	"github.com/ContainerSolutions/helm-convert/pkg/utils"
)

type labelsTransformer struct {
	keys []string
}

var _ Transformer = &labelsTransformer{}

// NewLabelsTransformer constructs a labelsTransformer.
func NewLabelsTransformer(keys []string) Transformer {
	return &labelsTransformer{keys}
}

// Transform finds common labels, if each resource contains a common label then
// the label is added to the kustomization.yaml file
func (t *labelsTransformer) Transform(config *ktypes.Kustomization, resources *types.Resources) error {
	// delete unwanted labels
	if err := t.removeLabels(config, resources); err != nil {
		return err
	}

	// retrieve common labels
	if err := t.commonLabels(config, resources); err != nil {
		return err
	}
	return nil
}

func (t *labelsTransformer) commonLabels(config *ktypes.Kustomization, resources *types.Resources) error {
	commonLabels := make(map[string]string, len(resources.ResMap))

	count := 0
RESOURCES_LOOP:
	for _, res := range resources.ResMap {
		obj := res.UnstructuredContent()

		if _, found := obj["metadata"]; !found {
			continue
		}

		metadata := obj["metadata"].(map[string]interface{})

		if _, found := metadata["labels"]; !found {
			continue
		}

		labels := metadata["labels"].(map[string]interface{})

		for key, value := range labels {
			if value == nil {
				continue RESOURCES_LOOP
			}

			labelValue := value.(string)
			if _, ok := commonLabels[key]; ok {
				if commonLabels[key] != labelValue {
					delete(commonLabels, key)
				}
			} else if count == 0 {
				commonLabels[key] = labelValue
			}
		}

		count++
	}

	if len(commonLabels) == 0 {
		return nil
	}

	// delete common labels from resources
	for _, res := range resources.ResMap {
		obj := res.UnstructuredContent()

		if _, found := obj["metadata"]; !found {
			continue
		}

		metadata := obj["metadata"].(map[string]interface{})

		if _, found := metadata["labels"]; !found {
			continue
		}

		labels := metadata["labels"].(map[string]interface{})

		for ck, cv := range commonLabels {
			for lk, lv := range labels {
				if ck == lk && cv == lv {
					delete(labels, ck)
				}
			}
		}
	}

	config.CommonLabels = commonLabels

	return nil
}

func (t *labelsTransformer) removeLabels(config *ktypes.Kustomization, resources *types.Resources) error {
	paths := []string{"matchLabels", "labels"}
	for _, res := range resources.ResMap {
		obj := res.UnstructuredContent()
		for _, path := range paths {
			for _, key := range t.keys {
				utils.RecursivelyRemoveKey(path, key, obj)
			}
		}
	}
	return nil
}

package transformers

import (
	"github.com/ContainerSolutions/helm-convert/pkg/types"
	"github.com/ContainerSolutions/helm-convert/pkg/utils"
	ktypes "sigs.k8s.io/kustomize/pkg/types"
)

type annotationsTransformer struct {
	keys []string
}

var _ Transformer = &annotationsTransformer{}

// NewAnnotationsTransformer constructs a annotationsTransformer.
func NewAnnotationsTransformer(keys []string) Transformer {
	return &annotationsTransformer{keys}
}

// Transform remove given annotations from manifests
func (t *annotationsTransformer) Transform(config *ktypes.Kustomization, resources *types.Resources) error {
	// TODO: retrieve common annotations for config.CommonAnnotations
RESOURCES_LOOP:
	for id := range resources.ResMap {
		obj := resources.ResMap[id].Map()

		for _, key := range t.keys {
			err := utils.RecursivelyRemoveKey("annotations", key, obj)
			if err != nil {
				continue RESOURCES_LOOP
			}
		}
	}

	return nil
}

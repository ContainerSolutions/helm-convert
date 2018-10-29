package transformers

import (
	ktypes "sigs.k8s.io/kustomize/pkg/types"

	"github.com/ContainerSolutions/helm-convert/pkg/types"
	"github.com/ContainerSolutions/helm-convert/pkg/utils"
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
	for _, res := range resources.ResMap {
		obj := res.UnstructuredContent()

		for _, key := range t.keys {
			err := utils.RecursivelyRemoveKey("annotations", key, obj)
			if err != nil {
				continue RESOURCES_LOOP
			}
		}
	}

	return nil
}

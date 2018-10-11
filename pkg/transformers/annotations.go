package transformers

import (
	"github.com/ContainerSolutions/helm-convert/pkg/utils"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/types"
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
func (t *annotationsTransformer) Transform(config *types.Kustomization, resources resmap.ResMap) error {
	// TODO: retrieve common annotations for config.CommonAnnotations
	for _, res := range resources {
		obj := res.UnstructuredContent()

		for _, key := range t.keys {
			err := utils.RecursivelyRemoveKey("annotations", key, obj)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

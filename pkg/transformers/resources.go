package transformers

import (
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/types"

	"github.com/ContainerSolutions/helm-convert/pkg/utils"
)

type resourcesTransformer struct{}

var _ Transformer = &resourcesTransformer{}

// NewResourcesTransformer constructs a resourcesTransformer.
func NewResourcesTransformer() Transformer {
	return &resourcesTransformer{}
}

// Transform retrieve all manifests name and store them as resources in the kustomization.yaml
func (t *resourcesTransformer) Transform(config *types.Kustomization, resources resmap.ResMap) error {
	for id, res := range resources {
		filename, err := utils.GetResourceFileName(id, res)
		if err != nil {
			return err
		}
		config.Resources = append(config.Resources, filename)
	}

	return nil
}

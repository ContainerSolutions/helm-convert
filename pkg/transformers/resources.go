package transformers

import (
	"github.com/ContainerSolutions/helm-convert/pkg/types"
	"github.com/ContainerSolutions/helm-convert/pkg/utils"
	ktypes "sigs.k8s.io/kustomize/pkg/types"
)

type resourcesTransformer struct{}

var _ Transformer = &resourcesTransformer{}

// NewResourcesTransformer constructs a resourcesTransformer.
func NewResourcesTransformer() Transformer {
	return &resourcesTransformer{}
}

// Transform retrieve all manifests name and store them as resources in the kustomization.yaml
func (t *resourcesTransformer) Transform(config *ktypes.Kustomization, resources *types.Resources) error {
	for id, res := range resources.ResMap {
		filename, err := utils.GetResourceFileName(id, res)
		if err != nil {
			return err
		}
		config.Resources = append(config.Resources, filename)
	}

	return nil
}

package transformers

import (
	"path"
	"sort"

	"github.com/ContainerSolutions/helm-convert/pkg/types"
	"github.com/ContainerSolutions/helm-convert/pkg/utils"
	ktypes "sigs.k8s.io/kustomize/pkg/types"
)

type resourcesTransformer struct{ resourcePrefix string }

var _ Transformer = &resourcesTransformer{}

// NewResourcesTransformer constructs a resourcesTransformer.
func NewResourcesTransformer(resourcePrefix string) Transformer {
	return &resourcesTransformer{resourcePrefix}
}

// Transform retrieve all manifests name and store them as resources in the kustomization.yaml
func (t *resourcesTransformer) Transform(config *ktypes.Kustomization, resources *types.Resources) error {
	for id, res := range resources.ResMap {
		filename, err := utils.GetResourceFileName(id, res)
		if err != nil {
			return err
		}
		filename = path.Join(t.resourcePrefix, filename)

		config.Resources = append(config.Resources, filename)
	}

	sort.Strings(config.Resources)

	return nil
}

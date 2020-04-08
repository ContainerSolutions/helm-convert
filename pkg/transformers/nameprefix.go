package transformers

import (
	"github.com/ContainerSolutions/helm-convert/pkg/types"
	"github.com/ContainerSolutions/helm-convert/pkg/utils"
	ktypes "sigs.k8s.io/kustomize/pkg/types"
)

type namePrefixTransformer struct {
	ignoredPrefix string
}

var _ Transformer = &namePrefixTransformer{}

// NewNamePrefixTransformer constructs a namePrefixTransformer.
func NewNamePrefixTransformer(ignoredPrefix string) Transformer {
	return &namePrefixTransformer{ignoredPrefix}
}

// Transform retrieve all resource name, if a prefix is detected, add it to the kustomization.yaml file
func (t *namePrefixTransformer) Transform(config *ktypes.Kustomization, resources *types.Resources) error {
	var resourceName []string
	for _, res := range resources.ResMap {
		name, err := res.GetFieldValue("metadata.name")
		if err != nil {
			continue
		}

		resourceName = append(resourceName, name)
	}

	prefix := utils.GetPrefix(resourceName)

	if prefix != t.ignoredPrefix {
		config.NamePrefix = prefix
	}

	return nil
}

package transformers

import (
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/types"

	"github.com/ContainerSolutions/helm-convert/pkg/utils"
)

type namePrefixTransformer struct{}

var _ Transformer = &namePrefixTransformer{}

// NewNamePrefixTransformer constructs a namePrefixTransformer.
func NewNamePrefixTransformer() Transformer {
	return &namePrefixTransformer{}
}

// Transform retrieve all resource name, if a prefix is detected, add it to the kustomization.yaml file
func (t *namePrefixTransformer) Transform(config *types.Kustomization, resources resmap.ResMap) error {
	var resourceName []string
	for _, res := range resources {
		name, err := res.GetFieldValue("metadata.name")
		if err != nil {
			return err
		}

		resourceName = append(resourceName, name)
	}

	prefix := utils.GetPrefix(resourceName)

	if prefix != "" {
		config.NamePrefix = prefix
	}

	return nil
}

package transformers

import (
	"github.com/ContainerSolutions/helm-convert/pkg/types"
	ktypes "sigs.k8s.io/kustomize/pkg/types"
)

type namespaceTransformer struct{}

var _ Transformer = &namespaceTransformer{}

// NewNamespaceTransformer constructs a namespaceTransformer.
func NewNamespaceTransformer() Transformer {
	return &namespaceTransformer{}
}

// Transform set the namespace if all resources have the same namespace
func (t *namespaceTransformer) Transform(config *ktypes.Kustomization, resources *types.Resources) error {
	var namespace string
	for _, res := range resources.ResMap {
		resNamespace, err := res.GetFieldValue("metadata.namespace")
		if err != nil {
			continue
		}

		if namespace != "" && namespace != resNamespace {
			return nil
		}

		namespace = resNamespace
	}

	if namespace != "" {
		// Delete the namespace key if it is globally set
		for _, res := range resources.ResMap {
			_, err := res.GetFieldValue("metadata.namespace")
			if err != nil {
				continue
			}

			obj := res.UnstructuredContent()
			metadata := obj["metadata"].(map[string]interface{})
			delete(metadata, "namespace")
		}

		config.Namespace = namespace
	}

	return nil
}

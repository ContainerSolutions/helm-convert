package transformers

import (
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/types"
)

type namespaceTransformer struct{}

var _ Transformer = &namespaceTransformer{}

// NewNamespaceTransformer constructs a namespaceTransformer.
func NewNamespaceTransformer() Transformer {
	return &namespaceTransformer{}
}

// Transform set the namespace if all resources have the same namespace
func (t *namespaceTransformer) Transform(config *types.Kustomization, resources resmap.ResMap) error {
	var namespace string
	for _, res := range resources {
		obj := res.UnstructuredContent()

		_, found := obj["metadata"]
		if !found {
			continue
		}

		metadata := obj["metadata"].(map[string]interface{})

		n, found := metadata["namespace"]
		if !found {
			continue
		}

		resNamespace := n.(string)

		if namespace != "" && namespace != resNamespace {
			return nil
		}

		namespace = resNamespace
	}

	if namespace != "" {
		// Delete the namespace key if it is globally set
		for _, res := range resources {
			obj := res.UnstructuredContent()

			_, found := obj["metadata"]
			if !found {
				continue
			}

			metadata := obj["metadata"].(map[string]interface{})

			_, found = metadata["namespace"]
			if !found {
				continue
			}

			delete(metadata, "namespace")
		}

		config.Namespace = namespace
	}

	return nil
}

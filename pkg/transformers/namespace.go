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
		for _, res := range resources {
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

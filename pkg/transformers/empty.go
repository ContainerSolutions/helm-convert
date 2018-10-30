package transformers

import (
	"github.com/ContainerSolutions/helm-convert/pkg/types"
	ktypes "sigs.k8s.io/kustomize/pkg/types"
)

type emptyTransformer struct{}

var _ Transformer = &emptyTransformer{}

// NewEmptyTransformer constructs an emptyTransformer
func NewEmptyTransformer() Transformer {
	return &emptyTransformer{}
}

// Transform remove empty maps from manifests (ie: empty labels, resources, etc.)
func (t *emptyTransformer) Transform(config *ktypes.Kustomization, resources *types.Resources) error {
	for id := range resources.ResMap {
		obj := resources.ResMap[id].Map()

		_, err := t.emptyRecursive(obj)
		if err != nil {
			continue
		}
	}

	return nil
}

func (t *emptyTransformer) emptyRecursive(obj map[string]interface{}) (bool, error) {
	for key := range obj {
		switch typedV := obj[key].(type) {
		case map[string]interface{}:
			if len(typedV) == 0 {
				delete(obj, key)
			} else {
				d, err := t.emptyRecursive(typedV)
				if err != nil {
					return false, err
				}
				if d {
					delete(obj, key)
				}
			}
		case []interface{}:
			for i := range typedV {
				item := typedV[i]
				typedItem, ok := item.(map[string]interface{})
				if ok {
					d, err := t.emptyRecursive(typedItem)
					if err != nil {
						return false, err
					}
					if d {
						delete(obj, key)
					}
				}
			}
		// Remove keys with nil value
		case nil:
			delete(obj, key)
		}
	}

	return len(obj) == 0, nil
}

// Package transformers transform resources
package transformers

import (
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/types"
)

// A Transformer modifies an instance of resmap.ResMap.
type Transformer interface {
	// Transform modifies data in the argument, e.g. gathering common labels to
	// resources that can be labelled.
	Transform(*types.Kustomization, resmap.ResMap) error
}

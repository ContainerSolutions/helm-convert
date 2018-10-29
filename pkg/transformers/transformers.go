// Package transformers transform resources
package transformers

import (
	"github.com/ContainerSolutions/helm-convert/pkg/types"
	ktypes "sigs.k8s.io/kustomize/pkg/types"
)

// A Transformer modifies an instance of resources.
type Transformer interface {
	// Transform modifies data in the argument, e.g. gathering common labels to
	// resources that can be labelled.
	Transform(*ktypes.Kustomization, *types.Resources) error
}

package types

import (
	"sigs.k8s.io/kustomize/pkg/resmap"
)

// Resources contains a list of resources
type Resources struct {
	// ResMap contains a list of Kustomize resources
	ResMap resmap.ResMap

	// SourceFiles contains a list of file retrieved from either configmaps or
	// secret resources. The key being the filename, and the value its content
	SourceFiles map[string]string
}

// NewResources constructs a new Resources
func NewResources() *Resources {
	return &Resources{
		ResMap:      resmap.ResMap{},
		SourceFiles: make(map[string]string),
	}
}

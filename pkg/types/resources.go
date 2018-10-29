package types

import (
	"sigs.k8s.io/kustomize/pkg/resmap"
)

// Resources contains a list of resources
type Resources struct {
	// ResMap contains a list of Kustomize resources
	ResMap resmap.ResMap

	// ConfigFiles contains a list of external configuration file retrieved from
	// configmaps resources, ie: a multiline application.properties
	ConfigFiles map[string]string
}

// NewResources constructs a new Resources
func NewResources() *Resources {
	return &Resources{
		ResMap:      resmap.ResMap{},
		ConfigFiles: make(map[string]string),
	}
}

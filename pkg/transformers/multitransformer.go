package transformers

import (
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/types"
)

// multiTransformer contains a list of transformers
type multiTransformer struct {
	transformers []Transformer
}

var _ Transformer = &multiTransformer{}

// NewMultiTransformer constructs a multiTransformer
func NewMultiTransformer(t []Transformer) Transformer {
	r := &multiTransformer{
		transformers: make([]Transformer, len(t)),
	}
	copy(r.transformers, t)
	return r
}

// Transform prepends the name prefix
func (o *multiTransformer) Transform(config *types.Kustomization, resources resmap.ResMap) error {
	for _, t := range o.transformers {
		err := t.Transform(config, resources)
		if err != nil {
			return err
		}
	}
	return nil
}

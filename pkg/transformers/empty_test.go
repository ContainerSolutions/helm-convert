package transformers

import (
	"fmt"
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/types"

	"github.com/davecgh/go-spew/spew"
)

type emptyTransformerArgs struct {
	config    *types.Kustomization
	resources resmap.ResMap
}

func TestEmptyRun(t *testing.T) {
	var ingress = gvk.Gvk{Kind: "Ingress"}
	var deploy = gvk.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"}

	for _, test := range []struct {
		name     string
		input    *emptyTransformerArgs
		expected *emptyTransformerArgs
	}{
		{
			name: "it should remove empty values",
			input: &emptyTransformerArgs{
				config: &types.Kustomization{},
				resources: resmap.ResMap{
					resource.NewResId(ingress, "ing1"): resource.NewResourceFromMap(
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Ingress",
							"metadata": map[string]interface{}{
								"name":   "ing1",
								"labels": map[string]interface{}{},
							},
						}),
					resource.NewResId(deploy, "deploy1"): resource.NewResourceFromMap(
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Deployment",
							"metadata": map[string]interface{}{
								"name": "deploy1",
								"labels": map[string]interface{}{
									"app": "deploy1",
								},
							},
							"spec": map[string]interface{}{
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"labels": map[string]interface{}{},
									},
								},
							},
						},
					),
					resource.NewResId(deploy, "deploy2"): resource.NewResourceFromMap(
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Deployment",
							"metadata": map[string]interface{}{
								"name": "deploy2",
								"labels": map[string]interface{}{
									"app": "deploy2",
								},
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":      "container-1",
										"resources": map[string]interface{}{},
									},
									map[string]interface{}{
										"name": "container-2",
										"resources": map[string]interface{}{
											"limits": map[string]interface{}{
												"cpu": "1",
											},
										},
									},
								},
							},
						}),
				},
			},
			expected: &emptyTransformerArgs{
				config: &types.Kustomization{},
				resources: resmap.ResMap{
					resource.NewResId(ingress, "ing1"): resource.NewResourceFromMap(
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Ingress",
							"metadata": map[string]interface{}{
								"name": "ing1",
							},
						}),
					resource.NewResId(deploy, "deploy1"): resource.NewResourceFromMap(
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Deployment",
							"metadata": map[string]interface{}{
								"name": "deploy1",
								"labels": map[string]interface{}{
									"app": "deploy1",
								},
							},
						}),
					resource.NewResId(deploy, "deploy2"): resource.NewResourceFromMap(
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Deployment",
							"metadata": map[string]interface{}{
								"name": "deploy2",
								"labels": map[string]interface{}{
									"app": "deploy2",
								},
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name": "container-1",
									},
									map[string]interface{}{
										"name": "container-2",
										"resources": map[string]interface{}{
											"limits": map[string]interface{}{
												"cpu": "1",
											},
										},
									},
								},
							},
						}),
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("%s", test.name), func(t *testing.T) {
			lt := NewEmptyTransformer()
			err := lt.Transform(test.input.config, test.input.resources)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(test.input.config, test.expected.config) {
				t.Fatalf(
					"expected: \n %v\ngot:\n %v",
					spew.Sdump(test.expected.config),
					spew.Sdump(test.input.config),
				)
			}

			if !reflect.DeepEqual(test.input.resources, test.expected.resources) {
				err = test.expected.resources.ErrorIfNotEqual(test.expected.resources)
				t.Fatalf(
					"expected: \n %v\ngot:\n %v",
					spew.Sdump(test.expected.resources),
					spew.Sdump(test.input.resources),
				)
			}
		})
	}
}

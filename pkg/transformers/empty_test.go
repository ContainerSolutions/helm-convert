package transformers

import (
	"fmt"
	"testing"

	"github.com/ContainerSolutions/helm-convert/pkg/types"
	"github.com/kylelemons/godebug/pretty"
	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	ktypes "sigs.k8s.io/kustomize/pkg/types"
)

type emptyTransformerArgs struct {
	config    *ktypes.Kustomization
	resources *types.Resources
}

func TestEmptyRun(t *testing.T) {
	var ingress = gvk.Gvk{Kind: "Ingress"}
	var deploy = gvk.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"}
	var rf = resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())

	for _, test := range []struct {
		name     string
		input    *emptyTransformerArgs
		expected *emptyTransformerArgs
	}{
		{
			name: "it should remove empty values",
			input: &emptyTransformerArgs{
				config: &ktypes.Kustomization{},
				resources: &types.Resources{
					ResMap: resmap.ResMap{
						resid.NewResId(ingress, "ing1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Ingress",
								"metadata": map[string]interface{}{
									"name":   "ing1",
									"labels": map[string]interface{}{},
								},
							}),
						resid.NewResId(deploy, "deploy1"): rf.FromMap(
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
						resid.NewResId(deploy, "deploy2"): rf.FromMap(
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
											"env": []interface{}{
												map[string]interface{}{
													"name":  "FOO",
													"value": "BAR",
												},
												map[string]interface{}{
													"name":  "FOO",
													"value": nil,
												},
											},
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
			expected: &emptyTransformerArgs{
				config: &ktypes.Kustomization{},
				resources: &types.Resources{
					ResMap: resmap.ResMap{
						resid.NewResId(ingress, "ing1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Ingress",
								"metadata": map[string]interface{}{
									"name": "ing1",
								},
							}),
						resid.NewResId(deploy, "deploy1"): rf.FromMap(
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
						resid.NewResId(deploy, "deploy2"): rf.FromMap(
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
											"env": []interface{}{
												map[string]interface{}{
													"name":  "FOO",
													"value": "BAR",
												},
												map[string]interface{}{
													"name": "FOO",
												},
											},
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
		},
	} {
		t.Run(fmt.Sprintf("%s", test.name), func(t *testing.T) {
			lt := NewEmptyTransformer()
			err := lt.Transform(test.input.config, test.input.resources)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := pretty.Compare(test.input.config, test.expected.config); diff != "" {
				t.Errorf("%s, diff: (-got +want)\n%s", test.name, diff)
			}

			if diff := pretty.Compare(test.input.resources, test.expected.resources); diff != "" {
				t.Errorf("%s, diff: (-got +want)\n%s", test.name, diff)
			}
		})
	}
}

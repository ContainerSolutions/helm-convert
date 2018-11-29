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

type labelsTransformerArgs struct {
	config    *ktypes.Kustomization
	resources *types.Resources
}

func TestLabelsRun(t *testing.T) {
	var service = gvk.Gvk{Version: "v1", Kind: "Service"}
	var cmap = gvk.Gvk{Version: "v1", Kind: "ConfigMap"}
	var deploy = gvk.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"}
	var rf = resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())

	for _, test := range []struct {
		name     string
		input    *labelsTransformerArgs
		expected *labelsTransformerArgs
	}{
		{
			name: "it should retrieve common labels",
			input: &labelsTransformerArgs{
				config: &ktypes.Kustomization{},
				resources: &types.Resources{
					ResMap: resmap.ResMap{
						resid.NewResId(cmap, "cm1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "ConfigMap",
								"metadata": map[string]interface{}{
									"name": "cm1",
									"labels": map[string]interface{}{
										"app":     "nginx",
										"version": "1.0.0",
									},
								},
							}),
						resid.NewResId(deploy, "deploy1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Deployment",
								"metadata": map[string]interface{}{
									"name": "deploy1",
									"labels": map[string]interface{}{
										"app":     "nginx",
										"version": "1.0.0",
									},
								},
							}),
						resid.NewResId(service, "service1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Service",
								"metadata": map[string]interface{}{
									"name": "service1",
									"labels": map[string]interface{}{
										"app":     "nginx",
										"version": "2.0.0",
									},
								},
							}),
					},
				},
			},
			expected: &labelsTransformerArgs{
				config: &ktypes.Kustomization{
					CommonLabels: map[string]string{
						"app": "nginx",
					},
				},
				resources: &types.Resources{
					ResMap: resmap.ResMap{
						resid.NewResId(cmap, "cm1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "ConfigMap",
								"metadata": map[string]interface{}{
									"name": "cm1",
									"labels": map[string]interface{}{
										"version": "1.0.0",
									},
								},
							}),
						resid.NewResId(deploy, "deploy1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Deployment",
								"metadata": map[string]interface{}{
									"name": "deploy1",
									"labels": map[string]interface{}{
										"version": "1.0.0",
									},
								},
							}),
						resid.NewResId(service, "service1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Service",
								"metadata": map[string]interface{}{
									"name": "service1",
									"labels": map[string]interface{}{
										"version": "2.0.0",
									},
								},
							}),
					},
				},
			},
		},
		{
			name: "it should not delete labels that are not shared across resources",
			input: &labelsTransformerArgs{
				config: &ktypes.Kustomization{},
				resources: &types.Resources{
					ResMap: resmap.ResMap{
						resid.NewResId(cmap, "cm1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "ConfigMap",
								"metadata": map[string]interface{}{
									"name": "cm1",
									"labels": map[string]interface{}{
										"app":     "nginx",
										"version": "1.0.0",
									},
								},
							}),
						resid.NewResId(deploy, "deploy1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Deployment",
								"metadata": map[string]interface{}{
									"name": "deploy1",
									"labels": map[string]interface{}{
										"app":     "my-app",
										"version": "1.0.0",
									},
								},
							}),
						resid.NewResId(service, "service1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Service",
								"metadata": map[string]interface{}{
									"name": "service1",
									"labels": map[string]interface{}{
										"app":     "nginx",
										"version": "2.0.0",
									},
								},
							}),
					},
				},
			},
			expected: &labelsTransformerArgs{
				config: &ktypes.Kustomization{},
				resources: &types.Resources{
					ResMap: resmap.ResMap{
						resid.NewResId(cmap, "cm1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "ConfigMap",
								"metadata": map[string]interface{}{
									"name": "cm1",
									"labels": map[string]interface{}{
										"app":     "nginx",
										"version": "1.0.0",
									},
								},
							}),
						resid.NewResId(deploy, "deploy1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Deployment",
								"metadata": map[string]interface{}{
									"name": "deploy1",
									"labels": map[string]interface{}{
										"app":     "my-app",
										"version": "1.0.0",
									},
								},
							}),
						resid.NewResId(service, "service1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Service",
								"metadata": map[string]interface{}{
									"name": "service1",
									"labels": map[string]interface{}{
										"app":     "nginx",
										"version": "2.0.0",
									},
								},
							}),
					},
				},
			},
		},
		{
			name: "it should remove helm labels",
			input: &labelsTransformerArgs{
				config: &ktypes.Kustomization{},
				resources: &types.Resources{
					ResMap: resmap.ResMap{
						resid.NewResId(cmap, "cm1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "ConfigMap",
								"metadata": map[string]interface{}{
									"name": "cm1",
									"labels": map[string]interface{}{
										"chart":    "nginx",
										"heritage": "Tiller",
										"release":  "nginx",
									},
								},
							}),
						resid.NewResId(deploy, "deploy1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Deployment",
								"metadata": map[string]interface{}{
									"name": "deploy1",
									"labels": map[string]interface{}{
										"chart":    "nginx",
										"heritage": "Tiller",
										"release":  "nginx",
									},
								},
							}),
						resid.NewResId(service, "service1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Service",
								"metadata": map[string]interface{}{
									"name": "service1",
									"labels": map[string]interface{}{
										"chart":    "nginx",
										"heritage": "Tiller",
										"release":  "nginx",
									},
								},
								"spec": map[string]interface{}{
									"selector": map[string]interface{}{
										"chart":    "nginx",
										"heritage": "Tiller",
										"release":  "nginx",
									},
								},
							}),
						resid.NewResId(service, "poddisruptionbudget1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "policy/v1beta1",
								"kind":       "PodDisruptionBudget",
								"metadata": map[string]interface{}{
									"name": "poddisruptionbudget1",
									"labels": map[string]interface{}{
										"chart":    "nginx",
										"heritage": "Tiller",
										"release":  "nginx",
									},
								},
								"spec": map[string]interface{}{
									"selector": map[string]interface{}{
										"matchLabels": map[string]interface{}{
											"chart":    "nginx",
											"heritage": "Tiller",
											"release":  "nginx",
										},
									},
								},
							}),
					},
				},
			},
			expected: &labelsTransformerArgs{
				config: &ktypes.Kustomization{},
				resources: &types.Resources{
					ResMap: resmap.ResMap{
						resid.NewResId(cmap, "cm1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "ConfigMap",
								"metadata": map[string]interface{}{
									"name":   "cm1",
									"labels": map[string]interface{}{},
								},
							}),
						resid.NewResId(deploy, "deploy1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Deployment",
								"metadata": map[string]interface{}{
									"name":   "deploy1",
									"labels": map[string]interface{}{},
								},
							}),
						resid.NewResId(service, "service1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Service",
								"metadata": map[string]interface{}{
									"name":   "service1",
									"labels": map[string]interface{}{},
								},
								"spec": map[string]interface{}{
									"selector": map[string]interface{}{},
								},
							}),
						resid.NewResId(service, "poddisruptionbudget1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "policy/v1beta1",
								"kind":       "PodDisruptionBudget",
								"metadata": map[string]interface{}{
									"name":   "poddisruptionbudget1",
									"labels": map[string]interface{}{},
								},
								"spec": map[string]interface{}{
									"selector": map[string]interface{}{
										"matchLabels": map[string]interface{}{},
									},
								},
							}),
					},
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("%s", test.name), func(t *testing.T) {
			lt := NewLabelsTransformer([]string{"chart", "release", "heritage"})
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

package transformers

import (
	"fmt"
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	ktypes "sigs.k8s.io/kustomize/pkg/types"

	"github.com/ContainerSolutions/helm-convert/pkg/types"
	"github.com/davecgh/go-spew/spew"
)

type labelsTransformerArgs struct {
	config    *ktypes.Kustomization
	resources *types.Resources
}

func TestLabelsRun(t *testing.T) {
	var service = gvk.Gvk{Version: "v1", Kind: "Service"}
	var cmap = gvk.Gvk{Version: "v1", Kind: "ConfigMap"}
	var deploy = gvk.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"}

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
						resource.NewResId(cmap, "cm1"): resource.NewResourceFromMap(
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
						resource.NewResId(deploy, "deploy1"): resource.NewResourceFromMap(
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
						resource.NewResId(service, "service1"): resource.NewResourceFromMap(
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
						resource.NewResId(cmap, "cm1"): resource.NewResourceFromMap(
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
						resource.NewResId(deploy, "deploy1"): resource.NewResourceFromMap(
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
						resource.NewResId(service, "service1"): resource.NewResourceFromMap(
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
						resource.NewResId(cmap, "cm1"): resource.NewResourceFromMap(
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
						resource.NewResId(deploy, "deploy1"): resource.NewResourceFromMap(
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
						resource.NewResId(service, "service1"): resource.NewResourceFromMap(
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
						resource.NewResId(cmap, "cm1"): resource.NewResourceFromMap(
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
						resource.NewResId(deploy, "deploy1"): resource.NewResourceFromMap(
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
						resource.NewResId(service, "service1"): resource.NewResourceFromMap(
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
	} {
		t.Run(fmt.Sprintf("%s", test.name), func(t *testing.T) {
			lt := NewLabelsTransformer([]string{})
			err := lt.Transform(test.input.config, test.input.resources)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(test.input.config, test.expected.config) {
				t.Fatalf(
					"expected: \n %v\ngot:\n %v",
					spew.Sdump(test.expected.config.CommonLabels),
					spew.Sdump(test.input.config.CommonLabels),
				)
			}

			if !reflect.DeepEqual(test.input.resources, test.expected.resources) {
				t.Fatalf(
					"expected: \n %v\ngot:\n %v",
					spew.Sdump(test.expected.resources),
					spew.Sdump(test.input.resources),
				)
			}
		})
	}
}

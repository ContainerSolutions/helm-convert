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

type namePrefixTransformerArgs struct {
	config    *types.Kustomization
	resources resmap.ResMap
}

func TestNamePrefixRun(t *testing.T) {
	var service = gvk.Gvk{Version: "v1", Kind: "Service"}
	var cmap = gvk.Gvk{Version: "v1", Kind: "ConfigMap"}
	var deploy = gvk.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"}

	for _, test := range []struct {
		name     string
		input    *namePrefixTransformerArgs
		expected *namePrefixTransformerArgs
	}{
		{
			name: "it should set the name prefix if it exists in the resource name",
			input: &namePrefixTransformerArgs{
				config: &types.Kustomization{},
				resources: resmap.ResMap{
					resource.NewResId(cmap, "cm1"): resource.NewResourceFromMap(
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name": "prefix-cm1",
							},
						}),
					resource.NewResId(deploy, "deploy1"): resource.NewResourceFromMap(
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Deployment",
							"metadata": map[string]interface{}{
								"name": "prefix-deploy1",
							},
						}),
					resource.NewResId(service, "service1"): resource.NewResourceFromMap(
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Service",
							"metadata": map[string]interface{}{
								"name": "prefix-service1",
							},
						}),
				},
			},
			expected: &namePrefixTransformerArgs{
				config: &types.Kustomization{
					NamePrefix: "prefix-",
				},
				resources: resmap.ResMap{
					resource.NewResId(cmap, "cm1"): resource.NewResourceFromMap(
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name": "prefix-cm1",
							},
						}),
					resource.NewResId(deploy, "deploy1"): resource.NewResourceFromMap(
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Deployment",
							"metadata": map[string]interface{}{
								"name": "prefix-deploy1",
							},
						}),
					resource.NewResId(service, "service1"): resource.NewResourceFromMap(
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Service",
							"metadata": map[string]interface{}{
								"name": "prefix-service1",
							},
						}),
				},
			},
		},
		{
			name: "it should not set the name prefix if there is no prefix detected in the resource name",
			input: &namePrefixTransformerArgs{
				config: &types.Kustomization{},
				resources: resmap.ResMap{
					resource.NewResId(cmap, "cm1"): resource.NewResourceFromMap(
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name": "prefix-cm1",
							},
						}),
					resource.NewResId(deploy, "deploy1"): resource.NewResourceFromMap(
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Deployment",
							"metadata": map[string]interface{}{
								"name": "deploy1",
							},
						}),
					resource.NewResId(service, "service1"): resource.NewResourceFromMap(
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Service",
							"metadata": map[string]interface{}{
								"name": "prefix-service1",
							},
						}),
				},
			},
			expected: &namePrefixTransformerArgs{
				config: &types.Kustomization{},
				resources: resmap.ResMap{
					resource.NewResId(cmap, "cm1"): resource.NewResourceFromMap(
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name": "prefix-cm1",
							},
						}),
					resource.NewResId(deploy, "deploy1"): resource.NewResourceFromMap(
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Deployment",
							"metadata": map[string]interface{}{
								"name": "deploy1",
							},
						}),
					resource.NewResId(service, "service1"): resource.NewResourceFromMap(
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Service",
							"metadata": map[string]interface{}{
								"name": "prefix-service1",
							},
						}),
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("%s", test.name), func(t *testing.T) {
			lt := NewNamePrefixTransformer()
			err := lt.Transform(test.input.config, test.input.resources)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(test.input.config, test.expected.config) {
				t.Fatalf(
					"expected: \n %v\ngot:\n %v",
					spew.Sdump(test.expected.config.NamePrefix),
					spew.Sdump(test.input.config.NamePrefix),
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

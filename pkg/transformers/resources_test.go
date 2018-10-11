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

type resourcesTransformerArgs struct {
	config    *types.Kustomization
	resources resmap.ResMap
}

func TestResourcesRun(t *testing.T) {
	var service = gvk.Gvk{Version: "v1", Kind: "Service"}
	var cmap = gvk.Gvk{Version: "v1", Kind: "ConfigMap"}
	var deploy = gvk.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"}

	for _, test := range []struct {
		name     string
		input    *resourcesTransformerArgs
		expected *resourcesTransformerArgs
	}{
		{
			name: "it should list all resources",
			input: &resourcesTransformerArgs{
				config: &types.Kustomization{},
				resources: resmap.ResMap{
					resource.NewResId(cmap, "cm1"): resource.NewResourceFromMap(
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name": "cm1",
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
								"name": "service1",
							},
						}),
				},
			},
			expected: &resourcesTransformerArgs{
				config: &types.Kustomization{
					Resources: []string{
						"cm1-cm.yaml",
						"deploy1-deploy.yaml",
						"service1-svc.yaml",
					},
				},
				resources: resmap.ResMap{
					resource.NewResId(cmap, "cm1"): resource.NewResourceFromMap(
						map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"name": "cm1",
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
								"name": "service1",
							},
						}),
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("%s", test.name), func(t *testing.T) {
			lt := NewResourcesTransformer()
			err := lt.Transform(test.input.config, test.input.resources)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(test.input.config, test.expected.config) {
				t.Fatalf(
					"expected: \n %v\ngot:\n %v",
					spew.Sdump(test.expected.config.Resources),
					spew.Sdump(test.input.config.Resources),
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

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for _, s1 := range a {
		ok := false
		for _, s2 := range b {
			if s1 == s2 {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}

	return true
}

// func deepEqual(a, b []interface{}) bool {
// 	if len(a) != len(b) {
// 		return false
// 	}

// 	for _, s1 := range a {
// 		ok := false
// 		for _, s2 := range b {
// 			if reflect.DeepEqual(s1, s2) {
// 				ok = true
// 				break
// 			}
// 		}
// 		if !ok {
// 			return false
// 		}
// 	}

// 	return true
// }

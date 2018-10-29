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

type namespaceTransformerArgs struct {
	config    *ktypes.Kustomization
	resources *types.Resources
}

func TestNamespaceRun(t *testing.T) {
	var service = gvk.Gvk{Version: "v1", Kind: "Service"}
	var cmap = gvk.Gvk{Version: "v1", Kind: "ConfigMap"}
	var deploy = gvk.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"}

	for _, test := range []struct {
		name     string
		input    *namespaceTransformerArgs
		expected *namespaceTransformerArgs
	}{
		{
			name: "it should set the namespace if all resource have a common namespace",
			input: &namespaceTransformerArgs{
				config: &ktypes.Kustomization{},
				resources: &types.Resources{
					ResMap: resmap.ResMap{
						resource.NewResId(cmap, "cm1"): resource.NewResourceFromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "ConfigMap",
								"metadata": map[string]interface{}{
									"name":      "cm1",
									"namespace": "staging",
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
									"name":      "service1",
									"namespace": "staging",
								},
							}),
					},
				},
			},
			expected: &namespaceTransformerArgs{
				config: &ktypes.Kustomization{
					Namespace: "staging",
				},
				resources: &types.Resources{
					ResMap: resmap.ResMap{
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
		},
		{
			name: "it should not set the namespace if all resource have different namespace",
			input: &namespaceTransformerArgs{
				config: &ktypes.Kustomization{},
				resources: &types.Resources{
					ResMap: resmap.ResMap{
						resource.NewResId(cmap, "cm1"): resource.NewResourceFromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "ConfigMap",
								"metadata": map[string]interface{}{
									"name":      "cm1",
									"namespace": "production",
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
									"name":      "service1",
									"namespace": "staging",
								},
							}),
					},
				},
			},
			expected: &namespaceTransformerArgs{
				config: &ktypes.Kustomization{},
				resources: &types.Resources{
					ResMap: resmap.ResMap{
						resource.NewResId(cmap, "cm1"): resource.NewResourceFromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "ConfigMap",
								"metadata": map[string]interface{}{
									"name":      "cm1",
									"namespace": "production",
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
									"name":      "service1",
									"namespace": "staging",
								},
							}),
					},
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("%s", test.name), func(t *testing.T) {
			lt := NewNamespaceTransformer()
			err := lt.Transform(test.input.config, test.input.resources)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(test.input.config, test.expected.config) {
				t.Fatalf(
					"expected: \n %v\ngot:\n %v",
					spew.Sdump(test.expected.config.Namespace),
					spew.Sdump(test.input.config.Namespace),
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

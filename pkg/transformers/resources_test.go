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

type resourcesTransformerArgs struct {
	config    *ktypes.Kustomization
	resources *types.Resources
}

func TestResourcesRun(t *testing.T) {
	var service = gvk.Gvk{Version: "v1", Kind: "Service"}
	var cmap = gvk.Gvk{Version: "v1", Kind: "ConfigMap"}
	var deploy = gvk.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"}
	var rf = resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())

	for _, test := range []struct {
		name     string
		input    *resourcesTransformerArgs
		expected *resourcesTransformerArgs
	}{
		{
			name: "it should list all resources",
			input: &resourcesTransformerArgs{
				config: &ktypes.Kustomization{},
				resources: &types.Resources{
					ResMap: resmap.ResMap{
						resid.NewResId(cmap, "cm1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "ConfigMap",
								"metadata": map[string]interface{}{
									"name": "cm1",
								},
							}),
						resid.NewResId(deploy, "deploy1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Deployment",
								"metadata": map[string]interface{}{
									"name": "deploy1",
								},
							}),
						resid.NewResId(service, "service1"): rf.FromMap(
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
			expected: &resourcesTransformerArgs{
				config: &ktypes.Kustomization{
					Resources: []string{
						"resources/cm1-cm.yaml",
						"resources/deploy1-deploy.yaml",
						"resources/service1-svc.yaml",
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
								},
							}),
						resid.NewResId(deploy, "deploy1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Deployment",
								"metadata": map[string]interface{}{
									"name": "deploy1",
								},
							}),
						resid.NewResId(service, "service1"): rf.FromMap(
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
	} {
		t.Run(fmt.Sprintf("%s", test.name), func(t *testing.T) {
			lt := NewResourcesTransformer()
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

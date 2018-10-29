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

type configMapTransformerArgs struct {
	config    *ktypes.Kustomization
	resources *types.Resources
}

func TestConfigMapRun(t *testing.T) {
	var configmap = gvk.Gvk{Version: "v1", Kind: "ConfigMap"}

	for _, test := range []struct {
		name     string
		input    *configMapTransformerArgs
		expected *configMapTransformerArgs
	}{
		{
			name: "it should convert configmaps",
			input: &configMapTransformerArgs{
				config: &ktypes.Kustomization{},
				resources: &types.Resources{
					ResMap: resmap.ResMap{
						resource.NewResId(configmap, "configmap1"): resource.NewResourceFromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "ConfigMap",
								"metadata": map[string]interface{}{
									"name": "configmap1",
								},
								"data": map[string]interface{}{
									"application.properties": `
app.name=My app
spring.jpa.hibernate.ddl-auto=update
spring.datasource.url=jdbc:mysql://<db_ip>:3306/db_example
spring.datasource.username=jane.doe
spring.datasource.password=pass123
`,
									"somekey":  "not a file",
									"SOME_ENV": "development",
								},
							}),
					},
					ConfigFiles: map[string]string{},
				},
			},
			expected: &configMapTransformerArgs{
				config: &ktypes.Kustomization{
					ConfigMapGenerator: []ktypes.ConfigMapArgs{
						ktypes.ConfigMapArgs{
							Name: "configmap1",
							DataSources: ktypes.DataSources{
								LiteralSources: []string{
									"somekey=\"not a file\"",
									"SOME_ENV=\"development\"",
								},
								FileSources: []string{"configmap1-application.properties"},
							},
						},
					},
				},
				resources: &types.Resources{
					ResMap: resmap.ResMap{},
					ConfigFiles: map[string]string{
						"configmap1-application.properties": `
app.name=My app
spring.jpa.hibernate.ddl-auto=update
spring.datasource.url=jdbc:mysql://<db_ip>:3306/db_example
spring.datasource.username=jane.doe
spring.datasource.password=pass123
`,
					},
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("%s", test.name), func(t *testing.T) {
			lt := NewConfigMapTransformer()
			err := lt.Transform(test.input.config, test.input.resources)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(test.input.config.ConfigMapGenerator, test.expected.config.ConfigMapGenerator) {
				t.Fatalf(
					"expected: \n %v\ngot:\n %v",
					spew.Sdump(test.expected.config.ConfigMapGenerator),
					spew.Sdump(test.input.config.ConfigMapGenerator),
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

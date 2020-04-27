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

type configMapTransformerArgs struct {
	config    *ktypes.Kustomization
	resources *types.Resources
}

func TestConfigMapRun(t *testing.T) {
	var configmap = gvk.Gvk{Version: "v1", Kind: "ConfigMap"}
	var rf = resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())

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
						resid.NewResId(configmap, "configmap1"): rf.FromMap(
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
					SourceFiles: map[string]string{},
				},
			},
			expected: &configMapTransformerArgs{
				config: &ktypes.Kustomization{
					ConfigMapGenerator: []ktypes.ConfigMapArgs{
						ktypes.ConfigMapArgs{
							GeneratorArgs: ktypes.GeneratorArgs{
								Name: "configmap1",
								DataSources: ktypes.DataSources{
									LiteralSources: []string{
										"SOME_ENV=development",
										"somekey=not a file",
									},
									FileSources: []string{"configmaps/configmap1/application.properties"},
								},
							},
						},
					},
				},
				resources: &types.Resources{
					ResMap: resmap.ResMap{},
					SourceFiles: map[string]string{
						"configmaps/configmap1/application.properties": `
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
			res := types.NewResources()
			res.ResMap = test.input.resources.ResMap

			lt := NewConfigMapTransformer()
			err := lt.Transform(test.input.config, res)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := pretty.Compare(test.input.config, test.expected.config); diff != "" {
				t.Errorf("%s, diff: (-got +want)\n%s", test.name, diff)
			}

			if diff := pretty.Compare(res, test.expected.resources); diff != "" {
				t.Errorf("%s, diff: (-got +want)\n%s", test.name, diff)
			}
		})
	}
}

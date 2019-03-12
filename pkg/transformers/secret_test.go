package transformers

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/ContainerSolutions/helm-convert/pkg/types"
	"github.com/kylelemons/godebug/pretty"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	ktypes "sigs.k8s.io/kustomize/pkg/types"
)

var rf = resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())

type secretTransformerArgs struct {
	config    *ktypes.Kustomization
	resources *types.Resources
}

func TestSecretRun(t *testing.T) {
	var secret = gvk.Gvk{Version: "v1", Kind: "Secret"}

	cert, err := ioutil.ReadFile("./testdata/tls.cert")
	if err != nil {
		t.Fatalf("Couldn't load tls.cert as test data")
	}
	key, err := ioutil.ReadFile("./testdata/tls.key")
	if err != nil {
		t.Fatalf("Couldn't load tls.key as test data")
	}

	for _, test := range []struct {
		name     string
		input    *secretTransformerArgs
		expected *secretTransformerArgs
	}{
		{
			name: "it should retrieve secrets",
			input: &secretTransformerArgs{
				config: &ktypes.Kustomization{},
				resources: &types.Resources{
					ResMap: resmap.ResMap{
						resid.NewResId(secret, "secret1"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Secret",
								"metadata": map[string]interface{}{
									"name": "secret1",
								},
								"type": string(corev1.SecretTypeOpaque),
								"data": map[string]interface{}{
									"DB_USERNAME": base64.StdEncoding.EncodeToString([]byte("admin")),
									"DB_PASSWORD": base64.StdEncoding.EncodeToString([]byte("password")),
								},
							}),
						resid.NewResId(secret, "secret2"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Secret",
								"metadata": map[string]interface{}{
									"name": "secret2",
								},
								"type": string(corev1.SecretTypeTLS),
								"data": map[string]interface{}{
									"tls.cert": base64.StdEncoding.EncodeToString(cert),
									"tls.key":  base64.StdEncoding.EncodeToString(key),
								},
							}),
						resid.NewResId(secret, "secret3"): rf.FromMap(
							map[string]interface{}{
								"apiVersion": "v1",
								"kind":       "Secret",
								"metadata": map[string]interface{}{
									"name": "secret3",
								},
								"type": string(corev1.SecretTypeOpaque),
								"data": nil,
							}),
					},
				},
			},
			expected: &secretTransformerArgs{
				config: &ktypes.Kustomization{
					SecretGenerator: []ktypes.SecretArgs{
						ktypes.SecretArgs{
							GeneratorArgs: ktypes.GeneratorArgs{
								Name: "secret1",
								DataSources: ktypes.DataSources{
									LiteralSources: []string{
										"DB_USERNAME=admin",
										"DB_PASSWORD=password",
									},
								},
							},
							Type: string(corev1.SecretTypeOpaque),
						},
						ktypes.SecretArgs{
							GeneratorArgs: ktypes.GeneratorArgs{
								Name: "secret2",
								DataSources: ktypes.DataSources{
									LiteralSources: []string{
										"tls.cert": string(cert),
										"tls.key":  string(key),
									},
								},
							},
							Type: string(corev1.SecretTypeTLS),
						},
						ktypes.SecretArgs{
							GeneratorArgs: ktypes.GeneratorArgs{
								Name: "secret3",
								DataSources: ktypes.DataSources{
									LiteralSources: []string{},
								},
							},
							Type: string(corev1.SecretTypeOpaque),
						},
					},
				,
				resources: &types.Resources{
					ResMap: resmap.ResMap{},
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("%s", test.name), func(t *testing.T) {
			lt := NewSecretTransformer()
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

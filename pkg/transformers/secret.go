package transformers

import (
	"encoding/base64"
	"fmt"
	"sort"

	"github.com/ContainerSolutions/helm-convert/pkg/types"
	ktypes "sigs.k8s.io/kustomize/pkg/types"
)

type secretTransformer struct{}

var _ Transformer = &secretTransformer{}

// NewSecretTransformer constructs a secretTransformer.
func NewSecretTransformer() Transformer {
	return &secretTransformer{}
}

// Transform retrieve secrets from manifests and store them as secretGenerator in the kustomization.yaml
func (t *secretTransformer) Transform(config *ktypes.Kustomization, resources *types.Resources) error {
	for id, res := range resources.ResMap {
		kind, err := res.GetFieldValue("kind")
		if err != nil {
			continue
		}

		if kind != "Secret" {
			continue
		}

		name, err := res.GetFieldValue("metadata.name")
		if err != nil {
			continue
		}

		secretType, err := res.GetFieldValue("type")
		if err != nil {
			secretType = "Opaque"
		}

		obj := resources.ResMap[id].Map()

		var data map[string]interface{}
		if _, found := obj["data"]; found && obj["data"] != nil {
			data = obj["data"].(map[string]interface{})
		}

		secretArg := ktypes.SecretArgs{
			GeneratorArgs: ktypes.GeneratorArgs{
				Name: name,
			},
			Type: secretType,
		}

		dataDecoded := make(map[string]string, len(data))
		for key, value := range data {
			decoded, err := base64.StdEncoding.DecodeString(value.(string))
			if err != nil {
				return fmt.Errorf("couldn't base64 decode the secret key '%s' with value '%v'", key, value)
			}
			dataDecoded[key] = string(decoded)
		}

		secretArg.GeneratorArgs.DataSources = TransformDataSource(t, name, dataDecoded, resources.SourceFiles)

		config.SecretGenerator = append(config.SecretGenerator, secretArg)
		delete(resources.ResMap, res.Id())
	}

	// sort by name
	sort.Slice(config.SecretGenerator, func(i, j int) bool {
		return config.SecretGenerator[i].Name < config.SecretGenerator[j].Name
	})

	return nil
}

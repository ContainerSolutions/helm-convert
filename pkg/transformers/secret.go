package transformers

import (
	"encoding/base64"
	"fmt"

	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/types"
)

type secretTransformer struct{}

var _ Transformer = &secretTransformer{}

// NewSecretTransformer constructs a secretTransformer.
func NewSecretTransformer() Transformer {
	return &secretTransformer{}
}

// Transform retrieve secrets from manifests and store them as secretGenerator in the kustomization.yaml
func (t *secretTransformer) Transform(config *types.Kustomization, resources resmap.ResMap) error {
	for _, res := range resources {
		kind, err := res.GetFieldValue("kind")
		if err != nil {
			return err
		}

		if kind != "Secret" {
			continue
		}

		name, err := res.GetFieldValue("metadata.name")
		if err != nil {
			return err
		}

		secretType, err := res.GetFieldValue("type")
		if err != nil {
			secretType = "Opaque"
		}

		obj := res.UnstructuredContent()

		_, found := obj["data"]
		if !found {
			return nil
		}

		data := obj["data"].(map[string]interface{})

		secretArg := types.SecretArgs{
			Name: name,
			Type: secretType,
		}

		commands := make(map[string]string)
		for key, value := range data {
			decoded, err := base64.StdEncoding.DecodeString(value.(string))
			if err != nil {
				return fmt.Errorf("couldn't base64 decode the secret key '%s' with value '%v'", key, value)
			}
			commands[string(key)] = fmt.Sprintf("printf \\\"%s\\\"", string(decoded))
		}

		secretArg.CommandSources = types.CommandSources{
			Commands: commands,
		}

		config.SecretGenerator = append(config.SecretGenerator, secretArg)
		delete(resources, res.Id())
	}

	return nil
}

// Package generators generate kustomize resources
package generators

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/ContainerSolutions/helm-convert/pkg/types"
	"github.com/ContainerSolutions/helm-convert/pkg/utils"
	"k8s.io/helm/pkg/proto/hapi/chart"
	ktypes "sigs.k8s.io/kustomize/pkg/types"
)

const (
	// DefaultKubeDescriptorFilename is the name of the kube-descriptor file
	DefaultKubeDescriptorFilename = "Kube-descriptor.yaml"

	// DefaultKustomizationFilename is the name of the kustomization config file
	DefaultKustomizationFilename = "kustomization.yaml"
)

// Generator type
type Generator struct {
	force bool
}

// NewGenerator contructs a new generator
func NewGenerator(force bool) *Generator {
	return &Generator{force}
}

// Render to disk the kustomization.yaml, Kube-descriptor.yaml and associated resources
func (g *Generator) Render(destination string, config *ktypes.Kustomization,
	metadata *chart.Metadata, resources *types.Resources, addConfigComments bool) error {
	var err error

	// chech if destination path already exist, prompt user to confirm override
	if ok, _ := utils.PathExists(destination); ok {
		if !g.force {
			reader := bufio.NewReader(os.Stdin)
			fmt.Printf("Destination directory '%s' already exist, override? [y/n] ", destination)
			approve, _ := reader.ReadString('\n')
			approve = strings.Trim(approve, " \n")

			if approve != "y" && approve != "yes" {
				return nil
			}
		}
	} else {
		os.MkdirAll(destination, os.ModePerm)
	}

	// render all manifests
	for id, res := range resources.ResMap {
		filename, err := utils.GetResourceFileName(id, res)
		if err != nil {
			return err
		}
		err = writeYamlFile(path.Join(destination, filename), res)
		if err != nil {
			return err
		}
	}

	// render all config and env files
	for filename, data := range resources.SourceFiles {
		// TODO: prevent overwriting of file, filename can be similar from one
		// resource to another
		err = writeFile(path.Join(destination, filename), []byte(data), 0644)
		if err != nil {
			return err
		}
	}

	// render kustomization.yaml
	err = writeYamlFile(path.Join(destination, DefaultKustomizationFilename), config)
	if err != nil {
		return err
	}

	// format and write kustomization.yaml
	err = writeAndFormatKustomizationConfig(path.Join(destination, DefaultKustomizationFilename), addConfigComments)
	if err != nil {
		return err
	}

	// render Kube-descriptor.yaml
	err = writeYamlFile(path.Join(destination, DefaultKubeDescriptorFilename), metadata)
	if err != nil {
		return err
	}

	return nil
}

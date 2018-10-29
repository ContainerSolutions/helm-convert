// Package generators generate kustomize resources
package generators

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	"k8s.io/helm/pkg/proto/hapi/chart"
	ktypes "sigs.k8s.io/kustomize/pkg/types"

	"github.com/ContainerSolutions/helm-convert/pkg/types"
	"github.com/ContainerSolutions/helm-convert/pkg/utils"
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
func (g *Generator) Render(destination string, config *ktypes.Kustomization, metadata *chart.Metadata, resources *types.Resources) error {
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

	// render all config files
	for filename, data := range resources.ConfigFiles {
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

	// render Kube-descriptor.yaml
	err = writeYamlFile(path.Join(destination, DefaultKubeDescriptorFilename), metadata)
	if err != nil {
		return err
	}

	return nil
}

// writeYamlFile write a given interface into yaml
func writeYamlFile(filePath string, data interface{}) error {
	output, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	return writeFile(filePath, output, 0644)
}

// writeFile writes data to a file named by filename.
func writeFile(filePath string, data []byte, perm os.FileMode) error {
	glog.V(4).Infof("Writing %s", filePath)

	err := ioutil.WriteFile(filePath, data, perm)
	if err != nil {
		return err
	}

	return nil
}

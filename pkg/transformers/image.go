package transformers

import (
	"sort"
	"strings"

	"github.com/ContainerSolutions/helm-convert/pkg/types"
	kimage "sigs.k8s.io/kustomize/pkg/image"
	ktypes "sigs.k8s.io/kustomize/pkg/types"
)

// imageTransformer replace images
type imageTransformer struct {
}

var _ Transformer = &imageTransformer{}

// NewImageTransformer constructs a imageTransformer.
func NewImageTransformer() Transformer {
	return &imageTransformer{}
}

// Transform finds all images and store them in the kustomization.yaml file
func (pt *imageTransformer) Transform(config *ktypes.Kustomization, resources *types.Resources) error {
	for id := range resources.ResMap {
		obj := resources.ResMap[id].Map()
		err := pt.findImage(config, obj)
		if err != nil {
			continue
		}
	}

	sort.Slice(config.Images, func(i, j int) bool {
		return imageString(config.Images[i]) < imageString(config.Images[j])
	})

	return nil
}

func (pt *imageTransformer) findImage(config *ktypes.Kustomization, obj map[string]interface{}) error {
	paths := []string{"containers", "initContainers"}
	found := false
	for _, path := range paths {
		val, found := obj[path]
		if found && val != nil {
			err := pt.getImageTag(config, obj, path)
			if err != nil {
				return err
			}
		}
	}
	if !found {
		return pt.findContainers(config, obj)
	}
	return nil
}

func (pt *imageTransformer) getImageTag(config *ktypes.Kustomization, obj map[string]interface{}, path string) error {
	containers := obj[path].([]interface{})
LOOP_CONTAINERS:
	for i := range containers {
		container := containers[i].(map[string]interface{})
		imagePath, found := container["image"]

		if !found {
			continue
		}

		imagePathStr := imagePath.(string)
		image := createKImage(imagePathStr)

		// don't add image if already in the list
		for _, v := range config.Images {
			if v.Name == image.Name {
				continue LOOP_CONTAINERS
			}
		}

		config.Images = append(config.Images, image)
	}
	return nil
}

func createKImage(imagePathStr string) kimage.Image {
	hasDigest := strings.Contains(imagePathStr, "@")
	separator := ":"

	if hasDigest {
		separator = "@"
	}

	s := strings.Split(imagePathStr, separator)
	image := kimage.Image{
		Name: s[0],
	}
	if len(s) > 1 {
		// combine everything but the last element in the string
		// fixes if image URL has a port
		image.Name = strings.Join(s[:len(s)-1], separator)
		if hasDigest {
			image.Digest = s[len(s)-1]
		} else {
			image.NewTag = s[len(s)-1]
		}
	}
	return image
}

func (pt *imageTransformer) findContainers(config *ktypes.Kustomization, obj map[string]interface{}) error {
	for key := range obj {
		switch typedV := obj[key].(type) {
		case map[string]interface{}:
			err := pt.findImage(config, typedV)
			if err != nil {
				return err
			}
		case []interface{}:
			for i := range typedV {
				item := typedV[i]
				typedItem, ok := item.(map[string]interface{})
				if ok {
					err := pt.findImage(config, typedItem)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func imageString(image kimage.Image) string {
	if image.Digest != "" {
		return image.Name + "@" + image.Digest
	}
	return image.Name + ":" + image.NewTag
}

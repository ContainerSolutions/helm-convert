package generators

import (
	"bufio"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"
)

// Pattern used to detect if a line contains a YAML key
var yamlKeyPattern = regexp.MustCompile("^[^ :]*:")

// writeYamlFile write a given interface into yaml
func writeYamlFile(filePath string, data interface{}) error {
	output, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	return writeFile(filePath, output, 0644)
}

// writeAndFormatKustomizationConfig adds line break and comments
func writeAndFormatKustomizationConfig(filePath string, comments bool) error {
	glog.V(4).Infof("Formatting %s", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	defer file.Close()

	var output []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// add line break before comment except for if this is the first line
		if yamlKeyPattern.MatchString(line) && len(output) > 0 {
			output = append(output, "")
		}

		// add comments
		if comments {
			for key, value := range commentsMapping {
				if strings.HasPrefix(line, key+":") {
					output = append(output, value)
				}
			}
		}

		output = append(output, line)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return writeFile(filePath, []byte(strings.Join(output, "\n")), 0644)
}

// writeFile writes data to a file named by filename.
func writeFile(filePath string, data []byte, perm os.FileMode) error {
	glog.V(4).Infof("Writing %s", filePath)

	os.MkdirAll(path.Dir(filePath), 0777)

	err := ioutil.WriteFile(filePath, data, perm)
	if err != nil {
		return err
	}

	return nil
}

package transformers

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/golang/glog"
	ktypes "sigs.k8s.io/kustomize/pkg/types"
)

var regexpEnv = regexp.MustCompile("^[A-Z0-9_]+$")

// TransformDataSource return a Kustomize DataSource from a given ConfigMap.Data or
// Secret.Data. If all keys from the resource matches an environment variable
// format, the resource is converted as EnvFile. If keys contains a file
// extension and the value is multiline then the file is stored as FileSources,
// otherwise LiteralSources.
func TransformDataSource(resourceName string, input map[string]string,
	sourceFiles map[string]string) (dataSources ktypes.DataSources) {

	if len(input) == 0 {
		return
	}

	if isEnvFile(input) {
		envFilename := fmt.Sprintf("%s.env", resourceName)
		envFile := TransformEnvDataSource(input)
		sourceFiles[envFilename] = envFile
		dataSources.EnvSource = envFilename

		glog.V(8).Infof("Converting '%s' as environment file with filename '%s'",
			resourceName, envFilename)
	} else {
		for key, value := range TransformFileDataSource(input) {
			filename := fmt.Sprintf("%s-%s", resourceName, key)
			sourceFiles[filename] = value
			dataSources.FileSources = append(dataSources.FileSources, filename)
		}
		dataSources.LiteralSources = TransformLiteralDataSource(input)

		sort.Strings(dataSources.FileSources)
		sort.Strings(dataSources.LiteralSources)

		glog.V(8).Infof("Converting %d file(s) as external file and %d literal(s) "+
			"from resource '%s'", len(dataSources.FileSources),
			len(dataSources.LiteralSources), resourceName)
	}

	return
}

// TransformEnvDataSource return an environment file from a given map
func TransformEnvDataSource(input map[string]string) (envFile string) {
	var envList []string
	for key, value := range input {
		if isEnvVariable(key) {
			envList = append(envList, fmt.Sprintf("%s=%s", key, value))
		}
	}
	sort.Strings(envList)
	envFile = strings.Join(envList, "\n")
	return
}

// TransformFileDataSource return a list of files from a given map
func TransformFileDataSource(input map[string]string) (files map[string]string) {
	files = make(map[string]string)
	for key, value := range input {
		if isFileExtension(key) && isMultiline(value) {
			files[key] = value
		}
	}
	return
}

// TransformLiteralDataSource return a list of literals (key=value) from a
// given map
func TransformLiteralDataSource(input map[string]string) (literal []string) {
	for key, value := range input {
		if !isFileExtension(key) && !isMultiline(value) {
			literal = append(literal, fmt.Sprintf("%s=%s", key, value))
		}
	}
	return
}

// isEnvFile return true if all the keys provided from a map match an
// environment variable pattern (uppercase, underscore separated words and value
// isn't multiline)
func isEnvFile(input map[string]string) bool {
	for key, value := range input {
		if !isEnvVariable(key) || isMultiline(value) {
			return false
		}
	}
	return true
}

// isMultiline return true if the provided value contains one of more line break
func isMultiline(s string) bool {
	return strings.Contains(s, "\n")
}

// isFileExtension return true if the provided string contains a dot
func isFileExtension(s string) bool {
	return strings.Contains(s, ".")
}

// isEnvVariable return true if the provided string match an environment
// variable pattern (uppercase, underscore separated words)
func isEnvVariable(key string) bool {
	return regexpEnv.MatchString(key)
}

// Package utils provide utilities functions to make helm-convert work
package utils

import (
	"fmt"
	"sort"
	"strings"

	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resource"
)

// GetResourceFileName return a resource name from metadata.name
func GetResourceFileName(id resid.ResId, res *resource.Resource) (string, error) {
	kind, err := res.GetFieldValue("kind")
	if err != nil {
		return "", err
	}

	name, err := res.GetFieldValue("metadata.name")
	if err != nil {
		return "", err
	}

	if strings.Contains(name, ":") {
		name = strings.ReplaceAll(name, ":", "-")
	}

	return strings.ToLower(fmt.Sprintf("resources/%s-%s.yaml", name, GetKindAbbreviation(kind))), nil
}

// GetKindAbbreviation return the abbreviation of a given resource
func GetKindAbbreviation(kind string) string {
	if abbrev, ok := K8SResourceMapping[strings.ToLower(kind)]; ok {
		return abbrev
	}
	return kind
}

// GetPrefix return the common prefix from a given list of string
func GetPrefix(s []string) string {
	sort.Sort(byLength(s))

	prefix := s[0]
	for _, name := range s[1:] {
		for i, s1 := range prefix {
			if string(s1) != string(name[i]) {
				prefix = prefix[0:i]
				break
			}
		}
	}

	return prefix
}

// RecursivelyRemoveKey of a matching key at a given path
func RecursivelyRemoveKey(path, key string, obj map[string]interface{}) error {
	for k := range obj {
		switch typedV := obj[k].(type) {
		case map[string]interface{}:
			if k == path {
				m := obj[path].(map[string]interface{})
				if _, exist := m[key]; exist {
					delete(m, key)
				}
			} else {
				err := RecursivelyRemoveKey(path, key, typedV)
				if err != nil {
					return err
				}
			}
		case []interface{}:
			for i := range typedV {
				item := typedV[i]
				typedItem, ok := item.(map[string]interface{})
				if ok {
					err := RecursivelyRemoveKey(path, key, typedItem)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

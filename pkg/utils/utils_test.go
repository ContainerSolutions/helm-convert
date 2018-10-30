package utils

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resource"
)

var rf = resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())

type getResourceFileNameArgs struct {
	id       resid.ResId
	resource *resource.Resource
}

func TestGetResourceFileName(t *testing.T) {
	var deploy = gvk.Gvk{Group: "apps", Version: "v1", Kind: "Deployment"}

	for _, test := range []struct {
		name     string
		input    getResourceFileNameArgs
		expected string
	}{
		{
			name: "it should return a filename",
			input: getResourceFileNameArgs{
				id: resid.NewResId(deploy, "deploy1"),
				resource: rf.FromMap(
					map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Deployment",
						"metadata": map[string]interface{}{
							"name": "my-deployment",
						},
					}),
			},
			expected: "my-deployment-deploy.yaml",
		},
	} {
		t.Run(fmt.Sprintf("%s", test.name), func(t *testing.T) {
			output, err := GetResourceFileName(test.input.id, test.input.resource)
			if err != nil {
				t.Fatalf("expected no error, got:\n %v", err)
			}
			if output != test.expected {
				t.Fatalf(
					"expected: \n %v\ngot:\n %v",
					test.expected,
					output,
				)
			}
		})
	}
}

func TestGetKindAbbreviation(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "it should return the resource abbreviation",
			input:    "deployment",
			expected: "deploy",
		},
		{
			name:     "it should return the full name when abbreviation doesn't exist",
			input:    "thisisanonexistingname",
			expected: "thisisanonexistingname",
		},
	} {
		t.Run(fmt.Sprintf("%s", test.name), func(t *testing.T) {
			output := GetKindAbbreviation(test.input)
			if output != test.expected {
				t.Fatalf(
					"expected: \n %v\ngot:\n %v",
					test.expected,
					test.input,
				)
			}
		})
	}
}

func TestGetPrefix(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name: "it should return a common prefix",
			input: []string{
				"prefix-deploy1",
				"prefix-service1",
				"prefix-cm1",
			},
			expected: "prefix-",
		},
		{
			name: "it should return an empty string if there is not common prefix",
			input: []string{
				"prefix-deploy1",
				"service1",
				"prefix-cm1",
			},
			expected: "",
		},
	} {
		t.Run(fmt.Sprintf("%s", test.name), func(t *testing.T) {
			output := GetPrefix(test.input)
			if output != test.expected {
				t.Fatalf(
					"expected: \n %v\ngot:\n %v",
					test.expected,
					test.input,
				)
			}
		})
	}
}

type recursivelyRemoveKeyArgs struct {
	path string
	key  string
	obj  map[string]interface{}
}

func TestRecursivelyRemoveKey(t *testing.T) {
	for _, test := range []struct {
		name     string
		input    *recursivelyRemoveKeyArgs
		expected map[string]interface{}
	}{
		{
			name: "it should recursively remove the key from the path",
			input: &recursivelyRemoveKeyArgs{
				path: "labels",
				key:  "label2",
				obj: map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"label1": "value1",
							"label2": "value2",
						},
					},
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"metadata": map[string]interface{}{
									"labels": map[string]interface{}{
										"label1": "value1",
										"label2": "value2",
									},
								},
							},
						},
					},
				},
			},
			expected: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"label1": "value1",
					},
				},
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"metadata": map[string]interface{}{
								"labels": map[string]interface{}{
									"label1": "value1",
								},
							},
						},
					},
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("%s", test.name), func(t *testing.T) {
			err := RecursivelyRemoveKey(test.input.path, test.input.key, test.input.obj)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(test.input.obj, test.expected) {
				t.Fatalf(
					"expected: \n %v\ngot:\n %v",
					spew.Sdump(test.expected),
					spew.Sdump(test.input.obj),
				)
			}
		})
	}
}

package transformers

import (
	"fmt"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	ktypes "sigs.k8s.io/kustomize/pkg/types"
)

func TestTransformDataSource(t *testing.T) {
	for _, test := range []struct {
		name               string
		resourceName       string
		input              map[string]string
		sourceFiles        map[string]string
		expectedSourceFile map[string]string
		expectedOutput     ktypes.DataSources
	}{
		{
			name:         "it should detect file source and literal",
			resourceName: "my-configmap",
			input: map[string]string{
				"somevar":  "single line",
				"name.txt": "multi\nline",
			},
			sourceFiles: map[string]string{
				"file1.yaml": "content",
			},
			expectedSourceFile: map[string]string{
				"file1.yaml":                       "content",
				"configmaps/my-configmap/name.txt": "multi\nline",
			},
			expectedOutput: ktypes.DataSources{
				LiteralSources: []string{
					"somevar=single line",
				},
				FileSources: []string{
					"configmaps/my-configmap/name.txt",
				},
			},
		},
		{
			name:         "it should detect env file",
			resourceName: "my-configmap",
			input: map[string]string{
				"NODE_ENV": "production",
				"SOMEENV":  "blop",
			},
			sourceFiles: map[string]string{
				"file1.yaml": "content",
			},
			expectedSourceFile: map[string]string{
				"file1.yaml":       "content",
				"my-configmap.env": "NODE_ENV=production\nSOMEENV=blop",
			},
			expectedOutput: ktypes.DataSources{
				EnvSource: "my-configmap.env",
			},
		},
	} {
		t.Run(fmt.Sprintf("%s", test.name), func(t *testing.T) {
			output := TransformDataSource(&configMapTransformer{}, test.resourceName, test.input, test.sourceFiles)
			if diff := pretty.Compare(output, test.expectedOutput); diff != "" {
				t.Errorf("%s, diff: (-got +want)\n%s", test.name, diff)
			}
			if diff := pretty.Compare(test.sourceFiles, test.expectedSourceFile); diff != "" {
				t.Errorf("%s, diff: (-got +want)\n%s", test.name, diff)
			}
		})
	}
}

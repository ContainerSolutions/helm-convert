package generators

var commentsMapping = map[string]string{
	"namespace": "# Adds namespace to all resources.",
	"namePrefix": "# Value of this field is prepended to the\n" +
		"# names of all resources",
	"commonLabels": "# Labels to add to all resources and selectors.",
	"commonAnnotations": "# Annotations (non-identifying metadata)\n" +
		"# to add to all resources. Like labels,\n" +
		"# these are key value pairs.",
	"resources": "# List of resource files that kustomize reads, modifies\n" +
		"# and emits as a YAML string",
	"configMapGenerator": "# Each entry in this list results in the creation of\n" +
		"# one ConfigMap resource (it's a generator of n maps).",
	"secretGenerator": "# Each entry in this list results in the creation of\n" +
		"# one Secret resource (it's a generator of n secrets).",
	"generatorOptions": "# generatorOptions modify behavior of all ConfigMap\n" +
		"# and Secret generators",
	"patches": "# Each entry in this list should resolve to\n" +
		"# a partial or complete resource definition file.",
	"patchesJson6902": "# Each entry in this list should resolve to\n" +
		"# a kubernetes object and a JSON patch that will be applied\n" +
		"# to the object.",
	"crds": "# Each entry in this list should be a relative path to\n" +
		"# a file for custom resource definition(CRD).",
	"vars": "# Vars are used to insert values from resources that cannot\n" +
		"# be referenced otherwise.",
	"imageTags": "# ImageTags modify the tags for images without\n" +
		"# creating patches.",
}

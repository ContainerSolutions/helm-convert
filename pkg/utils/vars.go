package utils

// K8SResourceMapping map long/short resource name
var K8SResourceMapping = map[string]string{
	"certificatesigningrequest": "csr",
	"clusterrolebinding":        "crb",
	"configmap":                 "cm",
	"customresourcedefinition":  "crd",
	"daemonset":                 "ds",
	"deployment":                "deploy",
	"endpoint":                  "ep",
	"horizontalpodautoscaler":   "hpa",
	"ingress":                   "ing",
	"limitrange":                "limits",
	"namespace":                 "ns",
	"networkpolicy":             "netpol",
	"persistentvolume":          "pv",
	"persistentvolumeclaim":     "pvc",
	"poddisruptionbudget":       "pdb",
	"podsecuritypolicy":         "psp",
	"replicaset":                "rs",
	"replicationcontroller":     "rc",
	"resourcequota":             "quota",
	"rolebinding":               "rb",
	"service":                   "svc",
	"serviceaccount":            "sa",
}

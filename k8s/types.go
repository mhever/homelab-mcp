package k8s

type NodeInfo struct {
	Name       string
	Status     string
	Roles      string
	Age        string
	Version    string
	InternalIP string
}

type PodInfo struct {
	Namespace string
	Name      string
	Ready     string // e.g. "2/3"
	Status    string
	Restarts  int
	Age       string
}

type EventInfo struct {
	Namespace string
	Type      string // Normal / Warning
	Reason    string
	Object    string // "kind/name"
	Message   string
	Age       string
}

type FluxInfo struct {
	Kind      string // "GitRepository" or "Kustomization"
	Namespace string
	Name      string
	Ready     string // condition status: "True", "False", "Unknown"
	Reason    string // condition reason
	Message   string
	Age       string
}

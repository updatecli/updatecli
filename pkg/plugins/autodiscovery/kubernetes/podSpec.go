package kubernetes

// kubernetesManifestSpec represents Kubernetes manifest
type kubernetesManifestSpec struct {
	ApiVersion string          `yaml:"apiVersion,omitempty"`
	Kind       string          `yaml:"kind,omitempty"`
	Spec       podTemplateSpec `yaml:"spec,omitempty"`
}

type podTemplateSpec struct {
	Containers []containerSpec  `yaml:"containers,omitempty"`
	Template   podTemplate2Spec `yaml:"template,omitempty"`
}

type podTemplate2Spec struct {
	Spec templateSpec `yaml:"spec,omitempty"`
}

type templateSpec struct {
	Containers []containerSpec `yaml:"containers,omitempty"`
}

type containerSpec struct {
	Name  string `yaml:"name,omitempty"`
	Image string `yaml:"image,omitempty"`
}

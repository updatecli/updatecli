package kubernetes

// kubernetesFlavorManifestSpec represents FlavorKubernetes manifest
type kubernetesFlavorManifestSpec struct {
	ApiVersion string          `yaml:"apiVersion,omitempty"`
	Kind       string          `yaml:"kind,omitempty"`
	Spec       podTemplateSpec `yaml:"spec,omitempty"`
}

type podTemplateSpec struct {
	Containers []containerSpec   `yaml:"containers,omitempty"`
	Template   podControllerSpec `yaml:"template,omitempty"`
}

type podControllerSpec struct {
	Spec templateSpec `yaml:"spec,omitempty"`
	Name string       `yaml:"name,omitempty"`
}

type templateSpec struct {
	Containers []containerSpec `yaml:"containers,omitempty"`
}

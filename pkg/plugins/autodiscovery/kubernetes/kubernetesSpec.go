package kubernetes

// kubernetesFlavorManifestSpec represents FlavorKubernetes manifest
type kubernetesFlavorManifestSpec struct {
	ApiVersion string          `yaml:"apiVersion,omitempty"`
	Kind       string          `yaml:"kind,omitempty"`
	Spec       podTemplateSpec `yaml:"spec,omitempty"`
}

type podTemplateSpec struct {
	InitContainers []containerSpec   `yaml:"initContainers,omitempty"`
	Containers     []containerSpec   `yaml:"containers,omitempty"`
	Template       podControllerSpec `yaml:"template,omitempty"`
	// cronJob
	JobTemplateSpec jobTemplateSpec `yaml:"jobTemplate,omitempty"`
}

type jobTemplateSpec struct {
	Spec struct {
		Template podControllerSpec `yaml:"template,omitempty"`
	} `yaml:"spec,omitempty"`
}

type podControllerSpec struct {
	Spec templateSpec `yaml:"spec,omitempty"`
	Name string       `yaml:"name,omitempty"`
}

type templateSpec struct {
	InitContainers []containerSpec `yaml:"initContainers,omitempty"`
	Containers     []containerSpec `yaml:"containers,omitempty"`
}

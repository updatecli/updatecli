package kubernetes

// kubernetesManifestSpec represents Kubernetes manifest
type kubernetesManifestSpec struct {
	ApiVersion string          `yaml:"apiVersion,omitempty"`
	Kind       string          `yaml:"kind,omitempty"`
	Spec       podTemplateSpec `yaml:"spec,omitempty"`
	// [ProwJobs](https://docs.prow.k8s.io/docs/jobs/) follows the Kubernetes spec
	ProwPreSubmitJobs  map[string][]podTemplate2Spec `yaml:"presubmits,omitempty"`
	ProwPostSubmitJobs map[string][]podTemplate2Spec `yaml:"postsubmits,omitempty"`
	ProwPeriodicJobs   []podTemplate2Spec            `yaml:"periodics,omitempty"`
}

type podTemplateSpec struct {
	Containers []containerSpec  `yaml:"containers,omitempty"`
	Template   podTemplate2Spec `yaml:"template,omitempty"`
}

type podTemplate2Spec struct {
	Spec templateSpec `yaml:"spec,omitempty"`
	Name string       `yaml:"name,omitempty"`
}

type templateSpec struct {
	Containers []containerSpec `yaml:"containers,omitempty"`
}

type containerSpec struct {
	Name  string `yaml:"name,omitempty"`
	Image string `yaml:"image,omitempty"`
}

package kubernetes

// kubernetesFlavourManifestSpec represents FlavourKubernetes manifest
type kubernetesFlavourManifestSpec struct {
	ApiVersion string          `yaml:"apiVersion,omitempty"`
	Kind       string          `yaml:"kind,omitempty"`
	Spec       podTemplateSpec `yaml:"spec,omitempty"`
}

// prowManifestSpec represents FlavourProw manifest
type prowFlavourManifestSpec struct {
	ProwPreSubmitJobs  map[string][]podControllerSpec `yaml:"presubmits,omitempty"`
	ProwPostSubmitJobs map[string][]podControllerSpec `yaml:"postsubmits,omitempty"`
	ProwPeriodicJobs   []podControllerSpec            `yaml:"periodics,omitempty"`
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

type containerSpec struct {
	Name  string `yaml:"name,omitempty"`
	Image string `yaml:"image,omitempty"`
}

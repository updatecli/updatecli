package kubernetes

// prowManifestSpec represents FlavorProw manifest
type prowFlavorManifestSpec struct {
	ProwPreSubmitJobs  map[string][]podControllerSpec `yaml:"presubmits,omitempty"`
	ProwPostSubmitJobs map[string][]podControllerSpec `yaml:"postsubmits,omitempty"`
	ProwPeriodicJobs   []podControllerSpec            `yaml:"periodics,omitempty"`
}

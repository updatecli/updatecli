package kubernetes

type containerSpec struct {
	Name  string `yaml:"name,omitempty"`
	Image string `yaml:"image,omitempty"`
}

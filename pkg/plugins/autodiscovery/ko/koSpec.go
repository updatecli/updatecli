package ko

// koSpec represents a ko data manifest
type koSpec struct {
	DefaultBaseImage   string            `yaml:"defaultBaseImage,omitempty"`
	BaseImageOverrides map[string]string `yaml:"baseImageOverrides,omitempty"`
}

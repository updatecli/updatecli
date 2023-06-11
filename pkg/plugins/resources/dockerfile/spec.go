package dockerfile

import (
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerfile/types"
)

// Spec defines a specification for a "dockerfile" resource
// parsed from an updatecli manifest file
type Spec struct {
	// File specifies the dockerimage file such as Dockerfile
	File string `yaml:"file,omitempty"`
	// Instruction specifies a DockerImage instruction such as ENV
	Instruction types.Instruction `yaml:"instruction,omitempty"`
	// Value specifies the value for a specified Dockerfile instruction.
	Value string `yaml:"value,omitempty"`
}

func (s Spec) Atomic() Spec {
	return Spec{
		File:        s.File,
		Instruction: s.Instruction,
		Value:       s.Value,
	}
}

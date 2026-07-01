package systemd

import (
	"fmt"
	"strings"
)

// Spec defines a specification for a "systemd" resource
// parsed from an updatecli manifest file
type Spec struct {
	// file specifies the systemd unit file path to manipulate
	//
	// compatible:
	//     * source
	//     * condition
	//     * target
	//
	// remark:
	//     * supports absolute or relative path
	//
	File string `yaml:",omitempty"`
	// section specifies the unit file section to interact with, such as "Unit", "Service",
	//
	// compatible:
	//     * source
	//     * condition
	//     * target
	//
	Section string `yaml:",omitempty"`
	// option specifies the key within the section to read or update, such as "ExecStart".
	//
	// compatible:
	//     * source
	//     * condition
	//     * target
	//
	Option string `yaml:",omitempty"`
	// index specifies which matching option to read or update when the same option is defined multiple times.
	// It starts at 0, so index 0 selects the first match, index 1 selects the second match, and so on.
	// If unset then a condition or a target matches every occurrences.
	//
	// compatible:
	//     * source
	//     * condition
	//     * target
	//
	// default:
	//     0
	//
	Index *int `yaml:",omitempty"`
	// value specifies the value for a specific option.
	//
	// compatible:
	//     * condition
	//     * target
	//
	// default:
	//     When used from a condition or a target, the default value is set to the associated source output.
	//
	Value string `yaml:",omitempty"`
}

func (s *Spec) Validate() error {
	var validationErrors []string

	if s.File == "" {
		validationErrors = append(validationErrors, "the attribute `spec.file` is required.")
	}

	if s.Section == "" {
		validationErrors = append(validationErrors, "the attribute `spec.section` is required.")
	}

	if s.Option == "" {
		validationErrors = append(validationErrors, "the attribute `spec.option` is required.")
	}

	if s.Index != nil && *s.Index < 0 {
		validationErrors = append(validationErrors, "the attribute `spec.index` must be greater than or equal to 0.")
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("validation error: the provided manifest configuration had the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	return nil
}

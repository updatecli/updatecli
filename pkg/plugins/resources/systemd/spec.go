package systemd

import (
	"fmt"
	"strings"
)

// Spec defines a specification for a "systemd" resource
// parsed from an updatecli manifest file
type Spec struct {
	//   `file` specifies the systemd unit file path to manipulate
	//
	//   compatible:
	//       * source
	//       * condition
	//       * target
	//
	//   remark:
	//       * supports absolute or relative path
	//
	File string `yaml:",omitempty"`
	//   `section` specifies the unit file section to interact with, such as "Unit", "Service",
	//   "Container", "Install", "Socket", "Timer", etc.
	//
	//   compatible:
	//       * source
	//       * condition
	//       * target
	//
	//   default:
	//       "Container"
	//
	Section string `yaml:",omitempty"`
	//   `option` specifies the key within the section to read or update,
	//   such as "Image", "ExecStart", "Environment", etc.
	//
	//   compatible:
	//       * source
	//       * condition
	//       * target
	//
	//   default:
	//       "Image"
	//
	Option string `yaml:",omitempty"`
	//   `value` specifies the value for a specific option.
	//
	//   compatible:
	//       * condition
	//       * target
	//
	//   default:
	//       When used from a condition or a target, the default value is set to the associated source output.
	//
	Value string `yaml:",omitempty"`
}

func (s *Spec) Validate() error {
	var validationErrors []string

	if s.File == "" {
		validationErrors = append(validationErrors, "Validation error in resource of type 'systemd': the attribute `spec.file` is required.")
	}

	if s.Section == "" {
		validationErrors = append(validationErrors, "Validation error in resource of type 'systemd': the attribute `spec.section` is required.")
	}

	if s.Option == "" {
		validationErrors = append(validationErrors, "Validation error in resource of type 'systemd': the attribute `spec.option` is required.")
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("validation error: the provided manifest configuration had the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	return nil
}

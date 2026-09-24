package systemd

import (
	"fmt"
	"strings"
)

/*
"systemd" defines the specification for manipulating systemd unit files.
It can be used as a "source", a "condition", or a "target".
*/
type Spec struct {
	// "file" defines the path of the systemd unit file to manipulate.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * "file" is required.
	//   * absolute and relative paths are supported.
	//
	// example:
	//   * file: /etc/systemd/system/myapp.service
	//
	File string `yaml:",omitempty"`
	// "section" defines the unit file section to use, such as "Unit" or "Service".
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * "section" is required.
	//
	// example:
	//   * section: Service
	//
	Section string `yaml:",omitempty"`
	// "option" defines the key within the section to read or update, such as "ExecStart".
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * "option" is required.
	//
	// example:
	//   * option: ExecStart
	//
	Option string `yaml:",omitempty"`
	// "index" defines which matching option to use when the same option is defined several times.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// default:
	//   empty
	//
	// remark:
	//   * it starts at 0, so index 0 selects the first match and index 1 the second.
	//   * it must be greater than or equal to 0.
	//   * when unset in a source, the first match is used.
	//   * when unset in a condition or a target, every match is used.
	//
	// example:
	//   * index: 0
	//
	Index *int `yaml:",omitempty"`
	// "value" defines the value of the option.
	//
	// compatible:
	//   * condition
	//   * target
	//
	// default:
	//   in a condition or a target, the output of the associated source.
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

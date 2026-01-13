package bazelmod

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines a specification for a "bazelmod" resource
// parsed from an updatecli manifest file
type Spec struct {
	// File specifies the path to the MODULE.bazel file
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// example:
	//   * MODULE.bazel
	//   * path/to/MODULE.bazel
	File string `yaml:",omitempty" jsonschema:"required"`
	// Module specifies the Bazel module name to target
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// example:
	//   * rules_go
	//   * gazelle
	//   * protobuf
	Module string `yaml:",omitempty" jsonschema:"required"`
	// VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	//
	// compatible:
	//   * source
	VersionFilter version.Filter `yaml:",omitempty"`
}

var (
	// ErrSpecFileUndefined is returned if a file wasn't specified
	ErrSpecFileUndefined = errors.New("bazelmod file undefined")
	// ErrSpecModuleUndefined is returned if a module wasn't specified
	ErrSpecModuleUndefined = errors.New("bazelmod module undefined")
	// ErrWrongSpec is returned when the Spec has wrong content
	ErrWrongSpec = errors.New("wrong spec content")
)

// Validate tests that the spec has the required fields
func (s *Spec) Validate() error {
	var errs []error

	if len(s.File) == 0 {
		errs = append(errs, ErrSpecFileUndefined)
	}
	if len(s.Module) == 0 {
		errs = append(errs, ErrSpecModuleUndefined)
	}

	for _, e := range errs {
		logrus.Errorln(e)
	}

	if len(errs) > 0 {
		return ErrWrongSpec
	}

	return nil
}

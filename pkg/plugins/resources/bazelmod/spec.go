package bazelmod

import (
	"errors"

	"github.com/sirupsen/logrus"
)

/*
"bazelmod" defines the specification for manipulating a module version in a Bazel MODULE.bazel file.
It can be used as a "source", a "condition", or a "target".
*/
type Spec struct {
	// "file" defines the path of the MODULE.bazel file.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * it is required.
	//   * a "file://" prefix is removed.
	//
	// example:
	//   * file: MODULE.bazel
	//   * file: path/to/MODULE.bazel
	//
	File string `yaml:",omitempty" jsonschema:"required"`
	// "module" defines the name of the Bazel module, as set in its "bazel_dep" entry.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * it is required.
	//
	// example:
	//   * module: rules_go
	//   * module: gazelle
	//   * module: protobuf
	//
	Module string `yaml:",omitempty" jsonschema:"required"`
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

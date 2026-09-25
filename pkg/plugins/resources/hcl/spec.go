package hcl

import (
	"errors"

	"github.com/sirupsen/logrus"
)

/*
"hcl" defines the specification for manipulating hcl files.
It can be used as a "source", a "condition", or a "target".
*/
type Spec struct {
	// "file" defines the path of the hcl file to use.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * "file" and "files" are mutually exclusive.
	//   * the schemes "https://", "http://" and "file://" are supported in a source or a condition.
	//
	// example:
	//   * file: main.tf
	//
	File string `yaml:",omitempty"`
	// "files" defines the list of hcl file paths to use.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * "file" and "files" are mutually exclusive.
	//   * in a source or a condition, "files" accepts only one entry.
	//   * the schemes "https://", "http://" and "file://" are supported in a source or a condition.
	//
	Files []string `yaml:",omitempty"`
	// "path" defines the hcl attribute path.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * "path" is required.
	//
	// example:
	//   * path: resource.aws_instance.app_server.ami
	//   * path: resource.helm_release.prometheus.version
	//   * path: plugin.aws.version
	//
	Path string `yaml:",omitempty"`
	// "value" defines the value associated with the hcl path.
	//
	// compatible:
	//   * condition
	//   * target
	//
	// default:
	//   the output of the associated source.
	//
	Value string `yaml:",omitempty"`
}

var (
	// ErrSpecFileUndefined is returned if a file wasn't specified
	ErrSpecFileUndefined = errors.New("hcl file undefined")
	// ErrSpecPathUndefined is returned if a path wasn't specified
	ErrSpecPathUndefined = errors.New("hcl path undefined")
	// ErrSpecFileAndFilesDefined when we both spec File and Files have been specified
	ErrSpecFileAndFilesDefined = errors.New("parameter \"file\" and \"files\" are mutually exclusive")
	// ErrWrongSpec is returned when the Spec has wrong content
	ErrWrongSpec error = errors.New("wrong spec content")
)

func (s *Spec) Validate() error {
	var errs []error

	if len(s.File) == 0 && len(s.Files) == 0 {
		errs = append(errs, ErrSpecFileUndefined)
	}
	if len(s.Path) == 0 {
		errs = append(errs, ErrSpecPathUndefined)
	}

	if len(s.File) > 0 && len(s.Files) > 0 {
		errs = append(errs, ErrSpecFileAndFilesDefined)
	}

	for _, e := range errs {
		logrus.Errorln(e)
	}

	if len(errs) > 0 {
		return ErrWrongSpec
	}

	return nil
}

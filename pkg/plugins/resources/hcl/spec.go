package hcl

import (
	"errors"

	"github.com/sirupsen/logrus"
)

/*
"hcl"  defines the specification for manipulating "hcl" files.
It can be used as a "source", a "condition", or a "target".
*/
type Spec struct {
	/*
		"file" defines the hcl file path to interact with.

		compatible:
			* source
			* condition
			* target

		remark:
			* "file" and "files" are mutually exclusive
			* protocols "https://", "http://", and "file://" are supported in path for source and condition
	*/
	File string `yaml:",omitempty"`
	/*
		"files" defines the list of hcl files path to interact with.

		compatible:
			* source
			* condition
			* target

		remark:
			* file and files are mutually exclusive
			* when using as a source only one file is supported
			* protocols "https://", "http://", and "file://" are supported in file path for source and condition
	*/
	Files []string `yaml:",omitempty"`
	/*
		"path" defines the hcl attribute path.

		compatible:
			* source
			* condition
			* target

		example:
			* path: resource.aws_instance.app_server.ami
			* path: resource.helm_release.prometheus.version
			* path: plugin.aws.version

	*/
	Path string `yaml:",omitempty"`
	/*
		"value" is the value associated with a hcl path.

		compatible:
			* condition
			* target

		default:
			When used from a condition or a target, the default value is set to linked source output.
	*/
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

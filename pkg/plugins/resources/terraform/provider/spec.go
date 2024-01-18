package provider

import (
	"errors"

	"github.com/sirupsen/logrus"
)

/*
"terraform/provider" defines the specification for manipulating providers in terraform files.
It can be used as a "condition", or a "target".
*/
type Spec struct {
	/*
		"file" defines the file path to interact with.

		compatible:
			* condition
			* target

		remark:
			* "file" and "files" are mutually exclusive
			* protocols "https://", "http://", and "file://" are supported in path for condition
	*/
	File string `yaml:",omitempty"`
	/*
		"files" defines the list of files path to interact with.

		compatible:
			* condition
			* target

		remark:
			* file and files are mutually exclusive
			* when using as a condition only one file is supported
			* protocols "https://", "http://", and "file://" are supported in file path for condition
	*/
	Files []string `yaml:",omitempty"`
	/*
		"value" is the value associated with a terraform provider.

		compatible:
			* condition
			* target

		default:
			When used from a condition or a target, the default value is set to linked source output.
	*/
	Value string `yaml:",omitempty"`

	/*
		"provider" is the terraform provider you wish to update.

		compatible:
			* condition
			* target
	*/
	Provider string `yaml:",omitempty"`
}

var (
	// ErrSpecFileUndefined is returned if a file wasn't specified
	ErrSpecFileUndefined = errors.New("terraform/provider file undefined")
	// ErrSpecProviderUndefined is returned if a provider wasn't specified
	ErrSpecProviderUndefined = errors.New("terraform/provider provider undefined")
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

	if len(s.File) > 0 && len(s.Files) > 0 {
		errs = append(errs, ErrSpecFileAndFilesDefined)
	}

	if len(s.Provider) == 0 {
		errs = append(errs, ErrSpecProviderUndefined)
	}

	for _, e := range errs {
		logrus.Errorln(e)
	}

	if len(errs) > 0 {
		return ErrWrongSpec
	}

	return nil
}

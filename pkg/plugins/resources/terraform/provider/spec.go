package provider

import (
	"errors"

	"github.com/sirupsen/logrus"
)

/*
"terraform/provider" defines the specification for manipulating providers in Terraform files.
It can be used as a "condition" or a "target".
*/
type Spec struct {
	// "file" defines the path of the Terraform file to use.
	//
	// compatible:
	//   * condition
	//   * target
	//
	// remark:
	//   * "file" and "files" are mutually exclusive.
	//   * the schemes "https://", "http://" and "file://" are supported in a condition.
	//
	File string `yaml:",omitempty"`
	// "files" defines the list of Terraform file paths to use.
	//
	// compatible:
	//   * condition
	//   * target
	//
	// remark:
	//   * "file" and "files" are mutually exclusive.
	//   * a condition only supports one file.
	//   * the schemes "https://", "http://" and "file://" are supported in a condition.
	//
	Files []string `yaml:",omitempty"`
	// "value" defines the version of the Terraform provider.
	//
	// compatible:
	//   * condition
	//   * target
	//
	// default:
	//   in a condition or a target, the output of the associated source.
	//
	Value string `yaml:",omitempty"`
	// "provider" defines the name of the Terraform provider to update, as declared in the "required_providers" block.
	//
	// compatible:
	//   * condition
	//   * target
	//
	// remark:
	//   * "provider" is required.
	//
	// example:
	//   * provider: kubernetes
	//
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

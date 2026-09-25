package lock

import (
	"errors"

	"github.com/sirupsen/logrus"
)

/*
"terraform/lock" defines the specification for manipulating .terraform.lock.hcl files.
It can be used as a "condition" or a "target".
*/
type Spec struct {
	// "file" defines the path of the Terraform lock file to use.
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
	// "files" defines the list of Terraform lock file paths to use.
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
	// "provider" defines the Terraform provider to update.
	//
	// compatible:
	//   * condition
	//   * target
	//
	// remark:
	//   * "provider" is required.
	//   * it accepts a provider address with or without the registry hostname.
	//
	// example:
	//   * provider: hashicorp/kubernetes
	//   * provider: registry.terraform.io/hashicorp/kubernetes
	//
	Provider string `yaml:",omitempty"`
	// "platforms" defines the target platforms to request package checksums for.
	//
	// compatible:
	//   * condition
	//   * target
	//
	// remark:
	//   * "platforms" is required.
	//
	// example:
	//   * platforms:
	//     - linux_amd64
	//     - darwin_arm64
	//
	Platforms []string `yaml:",omitempty"`
	// "skipconstraints" defines whether the constraints of the lock file are left untouched.
	//
	// compatible:
	//   * target
	//
	// default:
	//   false
	//
	// remark:
	//   * enabling it can break the lock file if the version from the source does not follow the constraints.
	//
	SkipConstraints bool `yaml:",omitempty"`
}

var (
	// ErrSpecFileUndefined is returned if a file wasn't specified
	ErrSpecFileUndefined = errors.New("terraform/lock file undefined")
	// ErrSpecProviderUndefined is returned if a provider wasn't specified
	ErrSpecProviderUndefined = errors.New("terraform/lock provider undefined")
	// ErrSpecPlatformsUndefined is returned if a platforms wasn't specified
	ErrSpecPlatformsUndefined = errors.New("terraform/lock platforms undefined")
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

	if len(s.Platforms) == 0 {
		errs = append(errs, ErrSpecPlatformsUndefined)
	}

	for _, e := range errs {
		logrus.Errorln(e)
	}

	if len(errs) > 0 {
		return ErrWrongSpec
	}

	return nil
}

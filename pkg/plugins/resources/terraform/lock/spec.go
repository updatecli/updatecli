package lock

import (
	"errors"

	"github.com/sirupsen/logrus"
)

/*
"terraform/lock"  defines the specification for manipulating .terraform-lock.hcl files.
It can be used as a "condition", or a "target".
*/
type Spec struct {
	/*
		"file" defines the terraform lock file path to interact with.

		compatible:
			* condition
			* target

		remark:
			* "file" and "files" are mutually exclusive
			* protocols "https://", "http://", and "file://" are supported in path for condition
	*/
	File string `yaml:",omitempty"`
	/*
		"files" defines the list of terraform lock files path to interact with.

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
		"provider" is the terraform provider you wish to update, supports with or without registry url.

		compatible:
			* condition
			* target
	*/
	Provider string `yaml:",omitempty"`

	/*
		"platforms" is the target platforms to request package checksums for.

		compatible:
			* condition
			* target
	*/
	Platforms []string `yaml:",omitempty"`

	/*
		"skipconstraints" will control whether the constraint in lock file is updated

		compatible:
			* condition
			* target

		NOTE: That turning this off can break the lockfile if version value source does not follow the constraints
	*/
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

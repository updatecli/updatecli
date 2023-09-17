package terraform

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines the Terraform parameters.
type Spec struct {
	// RootDir defines the root directory used to recursively search for .terraform.lock.hcl
	RootDir string `yaml:",omitempty"`
	// Ignore specifies rule to ignore .terraform.lock.hcl update.
	Ignore MatchingRules `yaml:",omitempty"`
	// Only specify required rule to restrict .terraform.lock.hcl update.
	Only MatchingRules `yaml:",omitempty"`
	/*
		`versionfilter` provides parameters to specify the version pattern to use when generating manifest.

		kind - semver
			versionfilter of kind `semver` uses semantic versioning as version filtering
			pattern accepts one of:
				`patch` - patch only update patch version
				`minor` - minor only update minor version
				`major` - major only update major versions
				`a version constraint` such as `>= 1.0.0`

		kind - regex
			versionfilter of kind `regex` uses regular expression as version filtering
			pattern accepts a valid regular expression

		example:
		```
			versionfilter:
				kind: semver
				pattern: minor
		```

		and its type like regex, semver, or just latest.
	*/
	VersionFilter version.Filter `yaml:",omitempty"`
	/*
		"platforms" is the target platforms to request package checksums for.
	*/
	Platforms []string `yaml:",omitempty"`
}

var (
	// ErrSpecPlatformsUndefined is returned if a platforms wasn't specified
	ErrSpecPlatformsUndefined = errors.New("terraform platforms undefined")
	// ErrWrongSpec is returned when the Spec has wrong content
	ErrWrongSpec error = errors.New("wrong spec content")
)

func (s *Spec) Validate() error {
	var errs []error

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

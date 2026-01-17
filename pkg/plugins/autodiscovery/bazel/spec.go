package bazel

import (
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines the Bazel autodiscovery parameters.
type Spec struct {
	// `rootdir` defines the root directory used to recursively search for `MODULE.bazel` files
	RootDir string `yaml:",omitempty"`
	// `ignore` specifies rule to ignore Bazel module updates.
	Ignore MatchingRules `yaml:",omitempty"`
	// `only` specify required rule to restrict Bazel module updates.
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
	*/
	VersionFilter version.Filter `yaml:",omitempty"`
}

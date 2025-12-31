package terragrunt

import (
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// GitHubSpec defines GitHub credentials for accessing private repositories
type GitHubSpec struct {
	// `token` specifies the GitHub token to use for authentication
	//
	// compatible:
	//   * autodiscovery
	//
	// default:
	//   automatically detected from UPDATECLI_GITHUB_TOKEN or GITHUB_TOKEN environment variables
	//
	// example:
	//   ghp_xxxxxxxxxxxx
	Token string `yaml:",omitempty"`
}

// Spec defines the Terraform parameters.
type Spec struct {
	// `rootdir` defines the root directory from where looking for terragrunt configuration
	RootDir string `yaml:",omitempty"`
	// `ignore` specifies rule to ignore `.terraform.lock.hcl` update.
	Ignore MatchingRules `yaml:",omitempty"`
	// `only` specify required rule to restrict `.terraform.lock.hcl` update.
	Only MatchingRules `yaml:",omitempty"`
	// `github` specifies the github credentials to use for accessing private repositories
	//
	// default:
	//   Token is automatically detected from UPDATECLI_GITHUB_TOKEN or GITHUB_TOKEN environment variables
	//
	// example:
	// ```
	//   github:
	//     token: "ghp_xxxxxxxxxxxx"
	// ```
	GitHub GitHubSpec `yaml:",omitempty"`
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
}

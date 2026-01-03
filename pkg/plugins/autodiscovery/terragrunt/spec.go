package terragrunt

import (
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines the Terraform parameters.
type Spec struct {
	// `rootdir` defines the root directory from where looking for terragrunt configuration
	RootDir string `yaml:",omitempty"`
	// `ignore` specifies rule to ignore `.terraform.lock.hcl` update.
	Ignore MatchingRules `yaml:",omitempty"`
	// `only` specify required rule to restrict `.terraform.lock.hcl` update.
	Only MatchingRules `yaml:",omitempty"`
	// `token` specifies the token to use for Git authentication when accessing private repositories.
	// Works with any Git provider (GitHub, GitLab, Bitbucket, Gitea, etc.)
	//
	// compatible:
	//   * autodiscovery
	//
	// default:
	//   When not specified: No authentication (suitable for public repositories)
	//
	// remark:
	//   Must be explicitly set for private repositories.
	//   Use template functions to read from environment: token: "{{ requiredEnv \"GITLAB_TOKEN\" }}"
	//
	// example:
	//   token: "ghp_xxxxxxxxxxxx"
	//   token: "glpat-xxxxxxxxxxxx"
	//   token: "{{ requiredEnv \"GITLAB_TOKEN\" }}"
	Token *string `yaml:",omitempty"`
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

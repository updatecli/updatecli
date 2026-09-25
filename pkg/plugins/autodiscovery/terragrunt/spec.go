package terragrunt

import (
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

/*
"terragrunt" defines the specification for the Terragrunt autodiscovery crawler.
It searches Terragrunt ".hcl" files and generates manifests to update the Terraform module versions they reference.
*/
type Spec struct {
	// "rootdir" defines the directory where the crawler starts searching for Terragrunt ".hcl" files.
	//
	// default:
	//   the scm directory when "scmid" is set, otherwise the directory relative paths resolve from, by default the working directory.
	//
	// remark:
	//   * a relative path is resolved from the default directory.
	//   * an absolute path is used as is, instead of the scm directory.
	//
	RootDir string `yaml:",omitempty"`
	// "ignore" defines rules to exclude matching Terraform modules from the autodiscovery.
	//
	// remark:
	//   * a Terraform module is ignored when it matches at least one rule.
	//
	Ignore MatchingRules `yaml:",omitempty"`
	// "only" defines rules to restrict the autodiscovery to matching Terraform modules.
	//
	// remark:
	//   * a Terraform module is kept only when it matches at least one rule.
	//
	Only MatchingRules `yaml:",omitempty"`
	// "token" defines the token used for Git authentication when accessing private module repositories.
	//
	// default:
	//   empty, no authentication, which suits public repositories.
	//
	// remark:
	//   * it works with any Git provider, such as GitHub, GitLab, Bitbucket or Gitea.
	//   * it must be set for private repositories.
	//   * it is only used for modules whose source is a Git repository.
	//   * use a template function to read it from the environment, such as `{{ requiredEnv "GITLAB_TOKEN" }}`.
	//
	// example:
	//   * token: "ghp_xxxxxxxxxxxx"
	//   * token: "glpat-xxxxxxxxxxxx"
	//   * token: "{{ requiredEnv \"GITLAB_TOKEN\" }}"
	//
	Token *string `yaml:",omitempty"`
	// "username" defines the username used for Git authentication when accessing private module repositories.
	//
	// default:
	//   "oauth2", which matches the GitHub scm plugin and is required for go-git HTTP basic authentication.
	//
	// remark:
	//   * it works with any Git provider, such as GitHub, GitLab, Bitbucket or Gitea.
	//   * it is only used when "token" is set.
	//   * with a token, the username is usually a placeholder since the token identifies the user.
	//   * common values are "oauth2", "x-access-token", "git", or a real username.
	//   * use a template function to read it from the environment, such as `{{ requiredEnv "GIT_USERNAME" }}`.
	//
	// example:
	//   * username: "git"
	//   * username: "oauth2"
	//   * username: "{{ requiredEnv \"GIT_USERNAME\" }}"
	//
	Username *string `yaml:",omitempty"`
	// "versionfilter" defines the version filter used by the generated manifests.
	//
	// default:
	//   kind "semver" with pattern "*", any version greater than or equal to the current one.
	//
	// remark:
	//   * with kind "semver", "pattern" accepts:
	//     * "prerelease": the latest prerelease of the current version.
	//     * "patch": patch updates only.
	//     * "minor": patch and minor updates.
	//     * "minoronly": minor updates only.
	//     * "major": patch, minor and major updates.
	//     * "majoronly": major updates only.
	//     * a version constraint, such as ">= 1.0.0".
	//   * with kind "regex", "pattern" accepts a regular expression.
	//   * more examples at https://www.updatecli.io/docs/core/versionfilter/
	//
	// example:
	//   ```
	//   versionfilter:
	//     kind: semver
	//     pattern: minor
	//   ```
	//
	VersionFilter version.Filter `yaml:",omitempty"`
}

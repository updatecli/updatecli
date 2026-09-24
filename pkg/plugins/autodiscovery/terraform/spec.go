package terraform

import (
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

/*
"terraform" defines the specification for the Terraform autodiscovery crawler.
It searches ".terraform.lock.hcl" files and generates manifests to update the providers they lock.
*/
type Spec struct {
	// "rootdir" defines the directory where the crawler starts searching for ".terraform.lock.hcl" files.
	//
	// default:
	//   the scm directory when "scmid" is set, otherwise the directory relative paths resolve from, by default the working directory.
	//
	// remark:
	//   * a relative path is resolved from the default directory.
	//   * an absolute path is used as is, instead of the scm directory.
	//
	RootDir string `yaml:",omitempty"`
	// "ignore" defines rules to exclude matching Terraform providers from the autodiscovery.
	//
	// remark:
	//   * a Terraform provider is ignored when it matches at least one rule.
	//
	Ignore MatchingRules `yaml:",omitempty"`
	// "only" defines rules to restrict the autodiscovery to matching Terraform providers.
	//
	// remark:
	//   * a Terraform provider is kept only when it matches at least one rule.
	//
	Only MatchingRules `yaml:",omitempty"`
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
	// "platforms" defines the target platforms to request provider checksums for.
	//
	// default:
	//   ```
	//   - linux_amd64
	//   - linux_arm64
	//   - darwin_amd64
	//   - darwin_arm64
	//   ```
	//
	Platforms []string `yaml:",omitempty"`
}

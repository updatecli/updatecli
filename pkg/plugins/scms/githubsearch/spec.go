package githubsearch

import (
	"errors"

	"github.com/updatecli/updatecli/pkg/plugins/scms/git/commit"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/sign"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github/app"
)

// Spec represents the configuration input
type Spec struct {
	// Limit defines the maximum number of repositories to return from the search query.
	//
	// compatible:
	//   * scm
	//
	// default:
	//   10
	//
	// remark:
	//   If limit is set to 0, all repositories matching the search query will be returned.
	//
	Limit *int `yaml:",omitempty"`
	// Search defines the GitHub repository search query.
	//
	// compatible:
	//   * scm
	//
	// remark:
	//   For more information about the search query syntax, please refer to the following documentation:
	//   https://docs.githubz.com/en/search-github/searching-on-github/searching-for-repositories
	//
	Search string `yaml:",omitempty" jsonschema:"required"`
	// "branch" defines the git branch to work on.
	//
	// compatible:
	//   * scm
	//
	// default:
	//   ^main$
	//
	// remark:
	//   depending on which resource references the GitHub scm, the behavior will be different.
	//
	//   If the scm is linked to a source or a condition (using scmid), the branch will be used to retrieve
	//   file(s) from that branch.
	//
	//   If the scm is linked to target then Updatecli creates a new "working branch" based on the branch value.
	//   The working branch created by Updatecli looks like "updatecli_<pipelineID>".
	//   The working branch can be disabled using the "workingBranch" parameter set to false.
	Branch string `yaml:",omitempty"`
	// WorkingBranchPrefix defines the prefix used to create a working branch.
	//
	// compatible:
	//   * scm
	//
	// default:
	//   updatecli
	//
	// remark:
	//   A working branch is composed of three components:
	//   1. WorkingBranchPrefix
	//   2. Target Branch
	//   3. PipelineID
	//
	//   If WorkingBranchPrefix is set to '', then
	//   the working branch will look like "<branch>_<pipelineID>".
	WorkingBranchPrefix *string `yaml:",omitempty"`
	// WorkingBranchSeparator defines the separator used to create a working branch.
	//
	// compatible:
	//   * scm
	//
	// default:
	//   "_"
	WorkingBranchSeparator *string `yaml:",omitempty"`
	// "directory" defines the local path where the git repository is cloned.
	//
	// compatible:
	//   * scm
	//
	// remark:
	//   Unless you know what you are doing, it is recommended to use the default value.
	//   The reason is that Updatecli may automatically clean up the directory after a pipeline execution.
	//
	// default:
	//   The default value is based on your local temporary directory like: (on Linux)
	//   /tmp/updatecli/github/<owner>/<repository>
	Directory string `yaml:",omitempty"`
	// "email" defines the email used to commit changes.
	//
	// compatible:
	//   * scm
	//
	// default:
	//   default set to your global git configuration
	Email string `yaml:",omitempty"`
	// "token" specifies the credential used to authenticate with GitHub API.
	//
	// compatible:
	//  * scm
	//
	// remark:
	//  A token is a sensitive information, it's recommended to not set this value directly in the configuration file
	//  but to use an environment variable or a SOPS file.
	//
	//  The value can be set to `{{ requiredEnv "GITHUB_TOKEN"}}` to retrieve the token from the environment variable `GITHUB_TOKEN`
	//
	//  or `{{ .github.token }}` to retrieve the token from a SOPS file.
	//  For more information, about a SOPS file, please refer to the following documentation:
	//  https://github.com/getsops/sops
	//
	Token string `yaml:",omitempty"`
	// "url" specifies the default github url in case of GitHub enterprise
	//
	// compatible:
	//   * scm
	//
	// default:
	//   github.com
	//
	URL string `yaml:",omitempty"`
	// "username" specifies the username used to authenticate with GitHub API.
	//
	// compatible:
	//   * scm
	//
	// remark:
	//  the token is usually enough to authenticate with GitHub API. Needed when working with GitHub private repositories.
	Username string `yaml:",omitempty"`
	// "user" specifies the user associated with new git commit messages created by Updatecli
	//
	// compatible:
	//  * scm
	User string `yaml:",omitempty"`
	// "gpg" specifies the GPG key and passphrased used for commit signing
	//
	// compatible:
	//   * scm
	GPG sign.GPGSpec `yaml:",omitempty"`
	// "force" is used during the git push phase to run `git push --force`.
	//
	// compatible:
	//   * scm
	//
	// default:
	//   false
	//
	// remark:
	//   When force is set to true, Updatecli also recreates the working branches that
	//   diverged from their base branch.
	Force *bool `yaml:",omitempty"`
	// "commitMessage" is used to generate the final commit message.
	//
	// compatible:
	//   * scm
	//
	// remark:
	//   it's worth mentioning that the commit message settings is applied to all targets linked to the same scm.
	CommitMessage commit.Commit `yaml:",omitempty"`
	// "submodules" defines if Updatecli should checkout submodules.
	//
	// compatible:
	//   * scm
	//
	// default: true
	Submodules *bool `yaml:",omitempty"`
	// "workingBranch" defines if Updatecli should use a temporary branch to work on.
	// If set to `true`, Updatecli create a temporary branch to work on, based on the branch value.
	//
	// compatible:
	//  * scm
	//
	// default: true
	WorkingBranch *bool `yaml:",omitempty"`
	// "commitUsingApi" defines if Updatecli should use GitHub GraphQL API to create the commit.
	// When set to `true`, a commit created from a GitHub action using the GITHUB_TOKEN will automatically be signed by GitHub.
	// More info on https://github.com/updatecli/updatecli/issues/1914
	//
	// compatible:
	//  * scm
	//
	// default: false
	CommitUsingAPI *bool `yaml:",omitempty"`
	// "app" specifies the GitHub App credentials used to authenticate with GitHub API.
	// It is not compatible with the "token" and "username" fields.
	// It is recommended to use the GitHub App authentication method for better security and granular permissions.
	// For more information, please refer to the following documentation:
	// https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/authenticating-as-a-github-app-installation
	App *app.Spec `yaml:",omitempty"`
}

// Validate validates the Spec fields.
func (s Spec) Validate() error {
	if s.Search == "" {
		return errors.New(ErrSearchQueryEmpty)
	}
	return nil
}

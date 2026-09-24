package githubsearch

import (
	"errors"

	"github.com/updatecli/updatecli/pkg/plugins/scms/git/commit"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/sign"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github/app"
)

/*
"githubsearch" defines the specification for searching GitHub repositories.
It generates one "github" scm for each matching branch of each repository returned by the search query.
*/
type Spec struct {
	// "depth" defines the depth used when cloning the git repository.
	//
	// default:
	//   empty, which means a full clone.
	//
	// remark:
	//   * a value greater than 0 creates a shallow clone, so Updatecli cannot see the full git history.
	//     Pushing changes may then fail, in which case setting "force" to true may be needed.
	//   * a negative value is rejected.
	//
	// example:
	//   * depth: 1
	//
	Depth *int `yaml:",omitempty"`
	// "limit" defines the maximum number of GitHub scm generated from the search results.
	//
	// default:
	//   10
	//
	// remark:
	//   * 0 means no limit.
	//   * one scm is generated for each matching branch of each repository,
	//     so the limit counts repository branches, not repositories.
	//
	Limit *int `yaml:",omitempty"`
	// "search" defines the GitHub repository search query.
	//
	// remark:
	//   * see https://docs.github.com/en/search-github/searching-on-github/searching-for-repositories
	//     for the query syntax.
	//
	// example:
	//   * search: org:updatecli topic:updatecli
	//
	Search string `yaml:",omitempty" jsonschema:"required"`
	// "branch" defines a regular expression matching the git branches to work on.
	//
	// default:
	//   ^main$
	//
	// remark:
	//   * one GitHub scm is generated for each matching branch of each discovered repository.
	//   * when a generated scm is used by a source or a condition, files are read from the matching branch.
	//   * when a generated scm is used by a target, Updatecli pushes changes to a working branch
	//     based on the matching branch, named "updatecli_<branch>_<pipelineid>" by default.
	//   * set "workingbranch" to false to push changes directly to the matching branch.
	//
	// example:
	//   * branch: ^main$
	//   * branch: ^release/.*$
	//
	Branch string `yaml:",omitempty"`
	// "workingbranchprefix" defines the prefix of the working branch name.
	//
	// default:
	//   updatecli
	//
	// remark:
	//   * the working branch name joins the prefix, the target branch and the pipeline ID,
	//     separated by "workingbranchseparator".
	//   * when set to an empty string, the name starts with the separator, for example "_main_<pipelineid>".
	//
	WorkingBranchPrefix *string `yaml:",omitempty"`
	// "workingbranchseparator" defines the separator between the parts of the working branch name.
	//
	// default:
	//   _
	//
	WorkingBranchSeparator *string `yaml:",omitempty"`
	// "directory" defines the local path where the git repository is cloned.
	//
	// default:
	//   a directory under the Updatecli temporary directory, such as
	//   "/tmp/updatecli/github/<owner>/<repository>" on Linux.
	//
	// remark:
	//   * keep the default value unless you have a good reason to change it,
	//     as Updatecli may delete the directory after a pipeline run.
	//   * the value is passed as is to every generated scm., so every repository then uses the same directory.
	//
	Directory string `yaml:",omitempty"`
	// "email" defines the email address used to author commits.
	//
	// default:
	//   updatecli-bot@updatecli.io
	//
	Email string `yaml:",omitempty"`
	// "token" defines the token used to authenticate with the GitHub API.
	//
	// remark:
	//   * "token" and "app" are mutually exclusive.
	//   * a token is sensitive, so avoid writing it in the manifest. Read it from an environment
	//     variable with `{{ requiredEnv "GITHUB_TOKEN" }}`, or from a SOPS file with `{{ .github.token }}`.
	//     See https://github.com/getsops/sops
	//   * the environment variable UPDATECLI_GITHUB_TOKEN, or the UPDATECLI_GITHUB_APP_* environment
	//     variables, take precedence over this value.
	//   * when no credential is set, Updatecli falls back to the environment variable GITHUB_TOKEN.
	//
	Token string `yaml:",omitempty"`
	// "url" defines the GitHub URL, to use a GitHub Enterprise instance.
	//
	// default:
	//   github.com
	//
	// remark:
	//   * the scheme "https://" is added when missing.
	//
	// example:
	//   * url: github.example.com
	//
	URL string `yaml:",omitempty"`
	// "username" defines the username used with the token to authenticate with the GitHub API.
	//
	// remark:
	//   * the token is usually enough on its own. A username may be needed for private repositories.
	//
	Username string `yaml:",omitempty"`
	// "user" defines the name used to author commits.
	//
	// default:
	//   updatecli-bot
	//
	User string `yaml:",omitempty"`
	// "gpg" defines the GPG key and passphrase used to sign commits.
	//
	GPG sign.GPGSpec `yaml:",omitempty"`
	// "force" defines whether Updatecli runs `git push --force` when pushing changes.
	//
	// default:
	//   true
	//
	// remark:
	//   * when true, Updatecli also recreates the working branches that diverged from their base branch.
	//   * when "workingbranch" is false and "force" is not set, each generated GitHub scm returns an error,
	//     to avoid force pushing to "branch" by mistake. Set "force" explicitly to confirm the behavior.
	//
	Force *bool `yaml:",omitempty"`
	// "commitmessage" defines the settings used to generate commit messages.
	//
	// remark:
	//   * the settings apply to every target using this scm.
	//
	CommitMessage commit.Commit `yaml:",omitempty"`
	// "submodules" defines whether Updatecli clones the git submodules of the repository.
	//
	// default:
	//   true
	//
	Submodules *bool `yaml:",omitempty"`
	// "workingbranch" defines whether Updatecli pushes changes to a temporary working branch
	// based on "branch", instead of pushing to "branch" directly.
	//
	// default:
	//   true
	//
	WorkingBranch *bool `yaml:",omitempty"`
	// "commitusingapi" defines whether Updatecli creates commits with the GitHub GraphQL API instead of git.
	//
	// default:
	//   false
	//
	// remark:
	//   * GitHub signs the commits created this way from a GitHub Actions workflow using the GITHUB_TOKEN.
	//     See https://github.com/updatecli/updatecli/issues/1914
	//
	CommitUsingAPI *bool `yaml:",omitempty"`
	// "app" defines the GitHub App credentials used to authenticate with the GitHub API.
	//
	// remark:
	//   * "app" and "token" are mutually exclusive, and "username" is ignored when "app" is set.
	//   * a GitHub App gives better security and finer permissions than a personal token.
	//   * see https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/authenticating-as-a-github-app-installation
	//
	App *app.Spec `yaml:",omitempty"`
}

// Validate validates the Spec fields.
func (s Spec) Validate() error {
	if s.Search == "" {
		return errors.New(ErrSearchQueryEmpty)
	}
	return nil
}

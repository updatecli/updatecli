package azuredevopssearch

import (
	"errors"
	"strings"

	azdoclient "github.com/updatecli/updatecli/pkg/plugins/resources/azuredevops/client"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/commit"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/sign"
)

const (
	DefaultRepositoryLimit = 10
	ErrOrganizationEmpty   = "azure DevOps organization is required for azuredevopssearch SCM"
)

/*
"azuredevopssearch" defines the specification for searching repositories in an Azure DevOps organization.
It generates one "azuredevops" scm for each matching branch of each matching repository.
*/
type Spec struct {
	// "organization" defines the Azure DevOps organization to search repositories in.
	//
	Organization string `yaml:",omitempty" jsonschema:"required"`
	// "url" defines the Azure DevOps base URL.
	//
	// default:
	//   https://dev.azure.com
	//
	// remark:
	//   * the scheme "https://" is added when missing.
	//
	URL string `yaml:",omitempty"`
	// "project" defines a regular expression matching the Azure DevOps projects to search in.
	//
	// default:
	//   .*
	//
	// example:
	//   * project: ^platform$
	//
	Project string `yaml:",omitempty"`
	// "repository" defines a regular expression matching the repositories to use.
	//
	// default:
	//   .*
	//
	// remark:
	//   * disabled repositories are always ignored.
	//
	// example:
	//   * repository: ^infra-.*$
	//
	Repository string `yaml:",omitempty"`
	// "limit" defines the maximum number of Azure DevOps scm generated from the matching repositories.
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
	// "branch" defines a regular expression matching the git branches to work on.
	//
	// default:
	//   ^main$
	//
	// remark:
	//   * one Azure DevOps scm is generated for each matching branch of each discovered repository.
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
	//   "/tmp/updatecli/azuredevops/<project>/<repository>" on Linux.
	//
	// remark:
	//   * keep the default value unless you have a good reason to change it,
	//     as Updatecli may delete the directory after a pipeline run.
	//   * the value is passed as is to every generated scm., so every repository then uses the same directory.
	//
	Directory string `yaml:",omitempty"`
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
	// "email" defines the email address used to author commits.
	//
	// default:
	//   updatecli-bot@updatecli.io
	//
	Email string `yaml:",omitempty"`
	// "token" defines the personal access token used to authenticate with Azure DevOps.
	//
	// remark:
	//   * a token is sensitive, so avoid writing it in the manifest.
	//   * the environment variable UPDATECLI_AZURE_DEVOPS_TOKEN takes precedence over this value.
	//
	Token string `yaml:",omitempty"`
	// "username" defines the username used for git authentication.
	//
	// remark:
	//   * the environment variable UPDATECLI_AZURE_DEVOPS_USERNAME takes precedence over this value.
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
}

// Validate validates the Spec fields.
func (s Spec) Validate() error {
	switch {
	case strings.TrimSpace(s.Organization) == "":
		return errors.New(ErrOrganizationEmpty)
	default:
		return nil
	}
}

func (s *Spec) sanitize() {
	s.Organization = strings.TrimSpace(s.Organization)
	s.Project = strings.TrimSpace(s.Project)
	s.Repository = strings.TrimSpace(s.Repository)
	s.URL = azdoclient.EnsureValidURL(s.URL)
}

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
	ErrProjectEmpty        = "azure DevOps project regex is required for azuredevopssearch SCM"
)

// Spec represents the configuration input for the azuredevopssearch SCM.
type Spec struct {
	// "organization" defines the Azure DevOps organization.
	Organization string `yaml:",omitempty" jsonschema:"required"`
	// "url" defines the Azure DevOps base URL.
	URL string `yaml:",omitempty"`
	// "project" defines the Azure DevOps project regex used to match projects to search in.
	Project string `yaml:",omitempty" jsonschema:"required"`
	// "repository" defines the Azure DevOps repository regex used to match repositories.
	Repository string `yaml:",omitempty"`
	// Limit defines the maximum number of repositories to return.
	Limit *int `yaml:",omitempty"`
	// "branch" defines the git branch regex to work on.
	Branch string `yaml:",omitempty"`
	// WorkingBranchPrefix defines the prefix used to create a working branch.
	WorkingBranchPrefix *string `yaml:",omitempty"`
	// WorkingBranchSeparator defines the separator used to create a working branch.
	WorkingBranchSeparator *string `yaml:",omitempty"`
	// "directory" defines the local path where the git repository is cloned.
	Directory string `yaml:",omitempty"`
	// Depth defines the depth used when cloning the git repository.
	Depth *int `yaml:",omitempty"`
	// "email" defines the email used to commit changes.
	Email string `yaml:",omitempty"`
	// "token" specifies the personal access token used to authenticate with Azure DevOps.
	Token string `yaml:",omitempty"`
	// "username" specifies the username used for git authentication.
	Username string `yaml:",omitempty"`
	// "user" specifies the user associated with new git commit messages created by Updatecli.
	User string `yaml:",omitempty"`
	// "gpg" specifies the GPG key and passphrase used for commit signing.
	GPG sign.GPGSpec `yaml:",omitempty"`
	// "force" is used during the git push phase to run `git push --force`.
	Force *bool `yaml:",omitempty"`
	// "commitMessage" is used to generate the final commit message.
	CommitMessage commit.Commit `yaml:",omitempty"`
	// "submodules" defines if Updatecli should checkout submodules.
	Submodules *bool `yaml:",omitempty"`
	// "workingBranch" defines if Updatecli should use a temporary branch to work on.
	WorkingBranch *bool `yaml:",omitempty"`
}

// Validate validates the Spec fields.
func (s Spec) Validate() error {
	switch {
	case strings.TrimSpace(s.Organization) == "":
		return errors.New(ErrOrganizationEmpty)
	case strings.TrimSpace(s.Project) == "":
		return errors.New(ErrProjectEmpty)
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

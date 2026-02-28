package gitlabsearch

import (
	"errors"

	"github.com/updatecli/updatecli/pkg/plugins/resources/gitlab/client"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/commit"
	"github.com/updatecli/updatecli/pkg/plugins/scms/git/sign"
)

// Spec represents the configuration input for the gitlabsearch SCM.
type Spec struct {
	client.Spec `yaml:",inline,omitempty"`
	// "group" defines the GitLab group (or subgroup) to search repositories in.
	//
	// compatible:
	//   * scm
	//
	// remark:
	//   Supports nested groups using slash notation, e.g. "myorg/myteam".
	//
	Group string `yaml:",omitempty" jsonschema:"required"`
	// "search" filters the repository list by name.
	//
	// compatible:
	//   * scm
	//
	// remark:
	//   When omitted, all projects in the group are returned.
	//
	Search string `yaml:",omitempty"`
	// "includeSubgroups" defines whether projects from subgroups should be included in the search results.
	//
	// compatible:
	//   * scm
	//
	// default:
	//   true
	//
	IncludeSubgroups *bool `yaml:",omitempty"`
	// Limit defines the maximum number of repositories to return.
	//
	// compatible:
	//   * scm
	//
	// default:
	//   10
	//
	// remark:
	//   If limit is set to 0, all repositories matching the query will be returned.
	//
	Limit *int `yaml:",omitempty"`
	// "branch" defines the git branch to work on.
	//
	// compatible:
	//   * scm
	//
	// default:
	//   ^main$
	//
	// remark:
	//   The branch value is a regular expression used to match branches across the discovered repositories.
	//
	//   If the scm is linked to a source or a condition (using scmid), the branch will be used to retrieve
	//   file(s) from that branch.
	//
	//   If the scm is linked to a target then Updatecli creates a new "working branch" based on the branch value.
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
	//   /tmp/updatecli/gitlab/<owner>/<repository>
	Directory string `yaml:",omitempty"`
	// "email" defines the email used to commit changes.
	//
	// compatible:
	//   * scm
	//
	// default:
	//   default set to your global git configuration
	Email string `yaml:",omitempty"`
	// "user" specifies the user associated with new git commit messages created by Updatecli
	//
	// compatible:
	//  * scm
	User string `yaml:",omitempty"`
	// "gpg" specifies the GPG key and passphrase used for commit signing
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
	//   true
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
}

// Validate validates the Spec fields.
func (s Spec) Validate() error {
	if s.Group == "" {
		return errors.New(ErrGroupEmpty)
	}
	return nil
}

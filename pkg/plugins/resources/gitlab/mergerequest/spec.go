package mergerequest

import (
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitlab/client"
)

/*
"gitlab/mergerequest" defines the specification for opening a GitLab merge request
with the changes made by the pipeline, or for updating the one already open.
*/
type Spec struct {
	client.Spec
	// "sourcebranch" defines the branch the merge request takes its changes from.
	//
	// default:
	//   the working branch of the associated scm of kind "gitlab".
	//
	// remark:
	//   * set it only when the default inherited from the scm does not fit.
	//
	SourceBranch string `yaml:",omitempty"`
	// "targetbranch" defines the branch the merge request is merged into.
	//
	// default:
	//   the branch of the associated scm of kind "gitlab".
	//
	// remark:
	//   * set it only when the default inherited from the scm does not fit.
	//   * the GitLab scm creates and uses a working branch such as updatecli_xxxx as the source branch.
	//
	TargetBranch string `yaml:",omitempty"`
	// "owner" defines the owner of the GitLab repository.
	//
	// default:
	//   the owner of the associated scm of kind "gitlab".
	//
	// remark:
	//   * set it only when the default inherited from the scm does not fit.
	//
	Owner string `yaml:",omitempty"`
	// "repository" defines the name of the GitLab repository, for a specific owner.
	//
	// default:
	//   the repository of the associated scm of kind "gitlab".
	//
	// remark:
	//   * set it only when the default inherited from the scm does not fit.
	//
	Repository string `yaml:",omitempty"`
	// "title" defines the title of the GitLab merge request.
	//
	// default:
	//   the title is taken from the first match of:
	//   1. "title" set in the action spec.
	//   2. "title" set in the action.
	//   3. the title of the first associated target.
	//   4. the pipeline title.
	//
	// remark:
	//   * setting "title" in the action is usually preferred.
	//
	// example:
	// ```
	// actions:
	//   default:
	//     kind: gitlab/mergerequest
	//     scmid: default
	//     title: This is my title
	// ```
	//
	Title string `yaml:",omitempty"`
	// "body" defines a custom merge request body.
	//
	// default:
	//   a body generated from the pipeline execution.
	//
	// remark:
	//   * the generated body is usually the right choice.
	//   * "body" is useful to add information for reviewers, such as a changelog url.
	//
	Body string `yaml:",omitempty"`
	// "assignees" defines the list of assignees to add to the merge request.
	//
	// default:
	//   empty
	//
	// remark:
	//   * only GitLab user IDs are accepted. To find a user ID:
	//     1. open the user's profile page.
	//     2. in the upper right corner, select Actions (or ⋮).
	//     3. select Copy user ID.
	//
	// example:
	//   * assignees: [123456]
	//
	Assignees []int64 `yaml:",omitempty"`
	// "reviewers" defines the list of reviewers to add to the merge request.
	//
	// default:
	//   empty
	//
	// remark:
	//   * only GitLab user IDs are accepted. To find a user ID:
	//     1. open the user's profile page.
	//     2. in the upper right corner, select Actions (or ⋮).
	//     3. select Copy user ID.
	//
	// example:
	//   * reviewers: [123456]
	//
	Reviewers []int64 `yaml:",omitempty"`
	// "squash" defines if all commits are squashed into a single commit on merge.
	//
	// default:
	//   false
	//
	// remark:
	//   * project settings might override this value.
	//
	Squash *bool `yaml:",omitempty"`
	// "removesourcebranch" defines if the source branch is removed when the merge request is merged.
	//
	// default:
	//   false
	//
	RemoveSourceBranch *bool `yaml:",omitempty"`
	// "labels" defines the labels of the merge request.
	//
	// default:
	//   empty
	//
	// remark:
	//   * a label that does not exist yet is created as a project label and assigned to the merge request.
	//
	Labels []string `yaml:",omitempty"`
	// "mergecommitmessage" defines the commit message used when the merge request is merged.
	//
	// default:
	//   empty
	//
	// remark:
	//   * when empty, GitLab uses the default message format defined in the project settings.
	//
	MergeCommitMessage *string `yaml:",omitempty"`
	// "squashcommitmessage" defines the commit message used when the merge request is squashed and merged.
	//
	// default:
	//   empty
	//
	// remark:
	//   * when empty, GitLab uses the default message format defined in the project settings.
	//
	SquashCommitMessage *string `yaml:",omitempty"`
	// "allowcollaboration" defines if members who can merge to the target branch may push commits to the source branch.
	//
	// default:
	//   false
	//
	// remark:
	//   * when true, members with write access to the repository can push commits
	//     to the source branch of the merge request.
	//
	AllowCollaboration *bool `yaml:",omitempty"`
	// "automerge" defines if the auto merge feature is enabled on a new merge request.
	//
	// default:
	//   false
	//
	// remark:
	//   * when true, the merge request is merged automatically once all conditions are met,
	//     such as a successful pipeline and the required approvals.
	//
	AutoMerge bool `yaml:",omitempty"`
}

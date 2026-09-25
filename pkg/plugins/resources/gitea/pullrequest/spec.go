package pullrequest

import (
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/client"
)

/*
"gitea/pullrequest" defines the specification for opening a pull request on a Gitea repository.
It is used by an action to propose the changes made by the pipeline targets.
*/
type Spec struct {
	client.Spec
	// "sourcebranch" defines the branch name used as the source of the Gitea pull request.
	//
	// default:
	//   the working branch of the scm, when the action uses a scm of kind "gitea".
	//
	// remark:
	//   * unless you know what you are doing, do not set this value and rely on the scm to provide it.
	//   * the Gitea scm creates and uses a working branch such as "updatecli_xxxx".
	//
	SourceBranch string `yaml:",inline,omitempty"`
	// "targetbranch" defines the branch name used as the target of the Gitea pull request.
	//
	// default:
	//   the branch of the scm, when the action uses a scm of kind "gitea".
	//
	// remark:
	//   * unless you know what you are doing, do not set this value and rely on the scm to provide it.
	//
	TargetBranch string `yaml:",inline,omitempty"`
	// "owner" defines the owner of the Gitea repository.
	//
	// default:
	//   the owner of the scm, when the action uses a scm of kind "gitea".
	//
	// remark:
	//   * unless you know what you are doing, do not set this value and rely on the scm to provide it.
	//
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// "repository" defines the name of the Gitea repository for a specific owner.
	//
	// default:
	//   the repository of the scm, when the action uses a scm of kind "gitea".
	//
	// remark:
	//   * unless you know what you are doing, do not set this value and rely on the scm to provide it.
	//
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// "title" defines the title of the Gitea pull request.
	//
	// default:
	//   the first match from the following list:
	//   1. the "title" set in the action spec.
	//   2. the "title" set in the action.
	//   3. the title of the first associated target.
	//   4. the title of the pipeline.
	//
	// remark:
	//   * setting the title in the action (option 2) is usually preferred.
	//
	// example:
	// ```
	//   actions:
	//     default:
	//       kind: gitea/pullrequest
	//       scmid: default
	//       title: This is my title
	// ```
	//
	Title string `yaml:",inline,omitempty"`
	// "body" defines a custom body for the pull request.
	//
	// default:
	//   a body generated from the pipeline execution.
	//
	// remark:
	//   * unless you know what you are doing, do not set this value and rely on the default.
	//   * it is useful to give reviewers additional information, such as a changelog url.
	//
	Body string `yaml:",inline,omitempty"`

	// "assignees" defines the list of users assigned to the pull request.
	//
	// default:
	//   no assignee.
	//
	// remark:
	//   * each user must have access to the repository.
	//
	// example:
	// ```
	//   assignees:
	//     - alice
	//     - bob
	// ```
	//
	Assignees []string `yaml:",omitempty"`
}

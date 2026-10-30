package pullrequest

import (
	"fmt"
	"strings"

	giteasdk "gitea.dev/sdk"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/client"
)

const (
	MergeStrategyAuto   = "auto"
	MergeStrategyManual = "manual"
)

// MergeSpec defines how Updatecli merges a pull request.
type MergeSpec struct {
	// "strategy" defines how the pull request is merged.
	//
	// default:
	//   manual
	//
	// remark:
	//   * accepted values are "auto", "manual" or empty, case-insensitive.
	//   * "manual" and empty neither request nor cancel auto-merge.
	//   * "auto" requests server-side auto-merge. Use Gitea 1.22+ or Forgejo 8+.
	//   * settings are applied when target changes trigger the action. Runs without target changes only report an existing PR.
	//   * configure required checks and branch protections on the repository. Merging may be immediate.
	//   * an existing auto-merge schedule is left unchanged.
	//   * auto-merge isn't requested on a draft pull request.
	//
	Strategy string `yaml:",omitempty"`
}

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
	//   * the title of an existing pull request is updated on each run.
	//     A draft keeps its "WIP:" or "[WIP]" prefix. With any other draft prefix, the title isn't updated.
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
	//   On an existing pull request, it's merged with the report in the current description.
	//   A description without a report, such as one written by hand, is replaced.
	//
	// remark:
	//   * unless you know what you are doing, do not set this value and rely on the default.
	//   * it is useful to give reviewers additional information, such as a changelog url.
	//   * when set, it also replaces the description of an existing pull request.
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
	// "merge" defines how Updatecli merges the pull request.
	//
	// default:
	//   strategy manual
	//
	// example:
	// ```
	//   merge:
	//     strategy: auto
	// ```
	//
	Merge MergeSpec `yaml:",omitempty"`
	// "mergemethod" defines the merge method used with "merge.strategy: auto".
	//
	// default:
	//   empty, which uses the repository's default merge style.
	//
	// remark:
	//   * accepted values are "merge", "rebase", "rebase-merge", "squash", "fast-forward-only" or empty, case-insensitive.
	//   * the server enforces which merge methods are allowed on the repository.
	//   * changing this value does not change the method of an existing auto-merge schedule.
	//
	MergeMethod string `yaml:",omitempty"`
}

// Validate ensures that the provided Spec is valid.
func (s *Spec) Validate() error {
	switch strings.ToLower(s.Merge.Strategy) {
	case "", MergeStrategyAuto, MergeStrategyManual:
	default:
		return fmt.Errorf("invalid merge strategy %q, accepting one of 'auto', 'manual', or ''", s.Merge.Strategy)
	}
	if s.MergeMethod != "" && !isMergeMethodValid(s.MergeMethod) {
		return fmt.Errorf("invalid merge method %q, accepting one of 'merge', 'rebase', 'rebase-merge', 'squash', 'fast-forward-only', or ''", s.MergeMethod)
	}
	return nil
}

func isMergeMethodValid(method string) bool {
	switch giteasdk.MergeStyle(strings.ToLower(method)) {
	case giteasdk.MergeStyleMerge, giteasdk.MergeStyleRebase, giteasdk.MergeStyleRebaseMerge,
		giteasdk.MergeStyleSquash, giteasdk.MergeStyleFastForwardOnly:
		return true
	}
	return false
}

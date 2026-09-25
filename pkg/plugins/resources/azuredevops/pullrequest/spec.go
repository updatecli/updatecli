package pullrequest

import azdoclient "github.com/updatecli/updatecli/pkg/plugins/resources/azuredevops/client"

/*
"azuredevops/pullrequest" defines the specification for opening or updating an Azure DevOps pull request.
*/
type Spec struct {
	azdoclient.Spec
	// "sourcebranch" defines the branch the pull request merges from.
	//
	// default:
	//   the working branch of the associated scm.
	//
	SourceBranch string `yaml:",omitempty"`
	// "targetbranch" defines the branch the pull request merges into.
	//
	// default:
	//   the branch of the associated scm.
	//
	TargetBranch string `yaml:",omitempty"`
	// "title" defines the pull request title.
	//
	// default:
	//   the action title.
	//
	Title string `yaml:",omitempty"`
	// "body" defines a custom pull request body.
	//
	// default:
	//   a report of the changes made by Updatecli.
	//
	// remark:
	//   * when set, it replaces the generated report.
	//
	Body string `yaml:",omitempty"`
	// "draft" defines whether the pull request is created as a draft.
	//
	// remark:
	//   * when unset, the Azure DevOps default applies and an existing pull request keeps its draft state.
	//
	// example:
	//   * draft: true
	//
	Draft *bool `yaml:",omitempty"`
}

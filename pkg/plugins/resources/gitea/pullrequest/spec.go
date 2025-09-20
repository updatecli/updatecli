package pullrequest

import (
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/client"
)

// Spec defines settings used to interact with Gitea pullrequest
// It's a mapping of user input from a Updatecli manifest and it shouldn't modified
type Spec struct {
	client.Spec
	/*
		"sourcebranch" defines the branch name used as a source to create the Gitea pullrequest.

		default:
			"sourcebranch" inherits the value from the scm branch if a scm of kind "gitea" is specified by the action.

		remark:
			unless you know what you are doing, you shouldn't set this value and rely on the scmid to provide the sane default.
	*/
	SourceBranch string `yaml:",inline,omitempty"`
	/*
		"targetbranch" defines the branch name used as a target to create the Gitea pullrequest.

		default:
			"targetbranch" inherits the value from the scm working branch if a scm of kind "gitea" is specified by the action.

		remark:
			unless you know what you are doing, you shouldn't set this value and rely on the scmid to provide the sane default.
			the Gitea scm will create and use a working branch such as updatecli_xxxx
	*/
	TargetBranch string `yaml:",inline,omitempty"`
	/*
		"owner" defines the Gitea repository owner.

		remark:
			unless you know what you are doing, you shouldn't set this value and rely on the scmid to provide the sane default.
	*/
	Owner string `yaml:",omitempty" jsonschema:"required"`
	/*
		"repository" defines the Gitea repository for a specific owner

		remark:
			unless you know what you are doing, you shouldn't set this value and rely on the scmid to provide the sane default.
	*/
	Repository string `yaml:",omitempty" jsonschema:"required"`
	/*
		"title" defines the Gitea pullrequest title

		default:
			A Gitea pullrequest title is defined by one of the following location (first match)
				1. title is defined by the spec such as:

					actions:
						default:
							kind: gitea/pullrequest
							scmid: default
							spec:
								title: This is my awesome title

				2. title is defined by the action such as:

					actions:
						default:
							kind: gitea/pullrequest
							scmid default
							title: This is my awesome title

				3. title is defined by the first associated target title

				4. title is defined by the pipeline title

		remark:
			usually we prefer to go with option 2
	*/
	Title string `yaml:",inline,omitempty"`
	/*
		"body" defines a custom body pullrequest.

		default:
			By default a pullrequest body is generated out of a pipeline execution.

		remark:
			Unless you know what you are doing, you shouldn't set this value and rely on the sane default.
			"body" is useful to provide additional information when reviewing pullrequest, such as changelog url.
	*/
	Body string `yaml:",inline,omitempty"`

	/*
		"assignees" defines a list of assignees for the pull request.

		default:
			No assignees are set on the pull request.

		remark:
			You can use this to assign specific users to review the pull request.
			Make sure the users you specify have access to the repository.
	*/
	Assignees []string `yaml:",omitempty"`
}

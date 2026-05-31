package pullrequest

import (
	"github.com/updatecli/updatecli/pkg/plugins/resources/tangled/client"
)

// Spec defines settings used to interact with Tangled pullrequest.
// It's a mapping of user input from a Updatecli manifest and it shouldn't be modified.
type Spec struct {
	client.Spec `yaml:",inline,omitempty"`
	/*
		"sourcebranch" defines the branch name used as a source to create the Tangled pullrequest.

		default:
			"sourcebranch" inherits the value from the scm branch if a scm of kind "tangled" is specified by the action.

		remark:
			unless you know what you are doing, you shouldn't set this value and rely on the scmid to provide the sane default.
	*/
	SourceBranch string `yaml:",inline,omitempty"`
	/*
		"targetbranch" defines the branch name used as a target to create the Tangled pullrequest.

		default:
			"targetbranch" inherits the value from the scm working branch if a scm of kind "tangled" is specified by the action.

		remark:
			unless you know what you are doing, you shouldn't set this value and rely on the scmid to provide the sane default.
	*/
	TargetBranch string `yaml:",inline,omitempty"`
	/*
		"knot" specifies the Tangled knot hosting the repository.

		remark:
			unless you know what you are doing, you shouldn't set this value and rely on the scmid to provide the sane default.
	*/
	Knot string `yaml:",omitempty"`
	/*
		"owner" specifies the Tangled repository owner handle (e.g. alice.tangled.sh).

		remark:
			unless you know what you are doing, you shouldn't set this value and rely on the scmid to provide the sane default.
	*/
	Owner string `yaml:",omitempty"`
	/*
		"repository" specifies the Tangled repository for a specific owner.

		remark:
			unless you know what you are doing, you shouldn't set this value and rely on the scmid to provide the sane default.
	*/
	Repository string `yaml:",omitempty"`
	/*
		"repoDid" optionally specifies the repository DID (decentralized identifier).

		remark:
			when set, this DID is used as the target repository of the pull request record
			instead of resolving the owner/repository pair to a DID via the appview.
	*/
	RepoDID string `yaml:",omitempty"`
	/*
		"title" defines the Tangled pullrequest title.

		default:
			Tangled pullrequest title is derived from the action title or pipeline title when omitted.
	*/
	Title string `yaml:",inline,omitempty"`
	/*
		"body" defines a custom pullrequest body.

		default:
			By default a pullrequest body is generated out of a pipeline execution.
	*/
	Body string `yaml:",inline,omitempty"`
}

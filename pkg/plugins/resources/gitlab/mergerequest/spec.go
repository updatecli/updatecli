package mergerequest

import (
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitlab/client"
)

// Spec defines settings used to interact with GitLab pullrequest
// It's a mapping of user input from a Updatecli manifest and it shouldn't modified
type Spec struct {
	client.Spec
	// "sourcebranch" defines the branch name used as a source to create the GitLab mergerequest.
	//
	// default:
	// 		"sourcebranch" inherits the value from the scm branch if a scm of kind "gitlab" is specified by the action.
	//
	// remark:
	// 		unless you know what you are doing, you shouldn't set this value and rely on the scmid to provide the sane default.
	SourceBranch string `yaml:",omitempty"`
	// "targetbranch" defines the branch name used as a target to create the GitLab mergerequest.
	//
	// default:
	// 		"targetbranch" inherits the value from the scm working branch if a scm of kind "gitlab" is specified by the action.
	//
	// remark:
	// 		unless you know what you are doing, you shouldn't set this value and rely on the scmid to provide the sane default.
	// 		the GitLab scm will create and use a working branch such as updatecli_xxxx
	TargetBranch string `yaml:",omitempty"`
	// "owner" defines the GitLab repository owner.
	//
	// remark:
	// 		unless you know what you are doing, you shouldn't set this value and rely on the scmid to provide the sane default.
	Owner string `yaml:",omitempty"`
	// "repository" defines the GitLab repository for a specific owner
	//
	// remark:
	// 		unless you know what you are doing, you shouldn't set this value and rely on the scmid to provide the sane default.
	Repository string `yaml:",omitempty"`
	// "title" defines the GitLab mergerequest title
	//
	// default:
	//	 	A GitLab mergerequest title is defined by one of the following location (first match)
	//	 		1. title is defined by the spec such as:
	//
	//	 			actions:
	//	 				default:
	//	 					kind: gitlab/mergerequest
	//	 					scmid: default
	//	 					spec:
	//	 						title: This is my awesome title
	//
	//	 		2. title is defined by the action such as:
	//
	//	 			actions:
	//	 				default:
	//	 					kind: gitlab/mergerequest
	//	 					scmid default
	//	 					title: This is my awesome title
	//
	//	 		3. title is defined by the first associated target title
	//
	//	 		4. title is defined by the pipeline title
	//
	// remark:
	// 		usually we prefer to go with option 2
	Title string `yaml:",omitempty"`
	// "body" defines a custom mergerequest body
	//
	// default:
	// 	By default a mergerequest body is generated out of a pipeline execution.
	//
	// remark:
	// 	Unless you know what you are doing, you shouldn't set this value and rely on the sane default.
	// 	"body" is useful to provide additional information when reviewing mergerequest, such as changelog url.
	Body string `yaml:",omitempty"`
	// "assignees" contains the list of assignee to add to the merge request
	//
	// default: empty
	//
	// remark:
	//		assignees only accept GitLab User IDs.
	// 		To find the user ID:
	// 			1. Go to the users’ profile page.
	// 			2. On the profile page, in the upper-right corner, select Actions (or ⋮).
	// 			3. Select Copy user ID.
	Assignees []int `yaml:",omitempty"`
	// "reviewers" contains the list of reviewers to add to the merge request
	//
	// default: empty
	//
	// remark:
	// 		assignees only accept GitLab User IDs.
	// 		To find the user ID:
	// 			1. Go to the users’ profile page.
	// 			2. On the profile page, in the upper-right corner, select Actions (or ⋮).
	// 			3. Select Copy user ID.
	Reviewers []int `yaml:",omitempty"`
	// "squash" defines if all commits should be squashed into a single commit on merge
	//
	// default: false
	//
	// remark:
	// 		project settings might override this value
	Squash bool `yaml:",omitempty"`
	// "removesourcebranch" is a flag indicating if a merge request should remove the source branch when merging
	//
	// default: false
	RemoveSourceBranch bool `yaml:",omitempty"`
	//
	// 	"labels" defines labels for the merge request.
	//
	// 	default: empty
	//
	// 	remark:
	// 		if a label does not already exist, this creates a new project label and assigns it to the merge request
	Labels []string `yaml:",omitempty"`
}

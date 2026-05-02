package pullrequest

import azdoclient "github.com/updatecli/updatecli/pkg/plugins/resources/azuredevops/client"

// Spec defines settings used to interact with Azure DevOps pull requests.
type Spec struct {
	azdoclient.Spec
	// "sourcebranch" defines the source branch used to create the pull request.
	SourceBranch string `yaml:",omitempty"`
	// "targetbranch" defines the target branch used to create the pull request.
	TargetBranch string `yaml:",omitempty"`
	// "title" defines the pull request title.
	Title string `yaml:",omitempty"`
	// "body" defines a custom pull request body.
	Body string `yaml:",omitempty"`
	// "draft" defines if the pull request should be created as draft.
	Draft *bool `yaml:",omitempty"`
}

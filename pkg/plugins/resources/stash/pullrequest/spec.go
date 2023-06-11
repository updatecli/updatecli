package pullrequest

import "github.com/updatecli/updatecli/pkg/plugins/resources/stash/client"

// Spec defines settings used to interact with Bitbucket pullrequest
// It's a mapping of user input from a Updatecli manifest and it shouldn't modified
type Spec struct {
	client.Spec
	// SourceBranch specifies the pullrequest source branch
	SourceBranch string `yaml:",inline,omitempty"`
	// TargetBranch specifies the pullrequest target branch
	TargetBranch string `yaml:",inline,omitempty"`
	// Owner specifies repository owner
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// Repository specifies the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// Title defines the Bitbucket pullrequest title.
	Title string `yaml:",inline,omitempty"`
	// Body defines the Bitbucket pullrequest body
	Body string `yaml:",inline,omitempty"`
}

func (s Spec) Atomic() Spec {
	return Spec{
		SourceBranch: s.SourceBranch,
		TargetBranch: s.TargetBranch,
		Owner:        s.Owner,
		Repository:   s.Repository,
	}
}

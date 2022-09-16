package pullrequest

import (
	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/client"

	giteascm "github.com/updatecli/updatecli/pkg/plugins/scms/gitea"
)

// Spec defines settings used to interact with Gitea pullrequest
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
	// Title defines the Gitea pullrequest title.
	Title string `yaml:",inline,omitempty"`
	// Body defines the Gitea pullrequest body
	Body string `yaml:",inline,omitempty"`
}

// Gitea contains information to interact with Gitea api
type Gitea struct {
	// spec contains inputs coming from updatecli configuration
	spec Spec
	// client handle the api authentication
	client client.Client
	// scm allows to interact with a scm object
	scm *giteascm.Gitea
	// SourceBranch specifies the pullrequest source branch.
	SourceBranch string `yaml:",inline,omitempty"`
	// TargetBranch specifies the pullrequest target branch
	TargetBranch string `yaml:",inline,omitempty"`
	// Owner specifies repository owner
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// Repository specifies the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
}

// New returns a new valid Gitea object.
func New(spec interface{}, scm *giteascm.Gitea) (Gitea, error) {

	var clientSpec client.Spec
	var s Spec

	// mapstructure.Decode cannot handle embedded fields
	// hence we decode it in two steps
	err := mapstructure.Decode(spec, &clientSpec)
	if err != nil {
		return Gitea{}, err
	}

	err = mapstructure.Decode(spec, &s)
	if err != nil {
		return Gitea{}, nil
	}

	if scm != nil {

		if len(clientSpec.Token) == 0 && len(scm.Spec.Token) > 0 {
			clientSpec.Token = scm.Spec.Token
		}

		if len(clientSpec.URL) == 0 && len(scm.Spec.URL) > 0 {
			clientSpec.URL = scm.Spec.URL
		}

		if len(clientSpec.Username) == 0 && len(scm.Spec.Username) > 0 {
			clientSpec.Username = scm.Spec.Username
		}
	}

	// Sanitize modifies the clientSpec so it must be done once initialization is completed
	err = clientSpec.Sanitize()
	if err != nil {
		return Gitea{}, err
	}

	c, err := client.New(clientSpec)

	if err != nil {
		return Gitea{}, err
	}

	g := Gitea{
		spec:   s,
		client: c,
		scm:    scm,
	}

	err = g.inheritFromScm()

	if err != nil {
		return Gitea{}, nil
	}

	return g, nil

}

package pullrequest

import (
	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/plugins/resources/azure/devops/client"

	azurescm "github.com/updatecli/updatecli/pkg/plugins/scms/azure"
)

// Spec defines settings used to interact with Azure DevOps pullrequest
// It's a mapping of user input from a Updatecli manifest and it shouldn't modified
type Spec struct {
	client.Spec
	// SourceBranch specifies the pullrequest source branch
	SourceBranch string `yaml:",inline,omitempty"`
	// TargetBranch specifies the pullrequest target branch
	TargetBranch string `yaml:",inline,omitempty"`
	// Title defines the Azure DevOps pullrequest title.
	Title string `yaml:",inline,omitempty"`
	// Body defines the Azure DevOps pullrequest body
	Body string `yaml:",inline,omitempty"`
}

// Azure contains information to interact with Azure api
type Azure struct {
	// spec contains inputs coming from updatecli configuration
	spec Spec
	// client handle the api authentication
	client client.Client
	// scm allows to interact with a scm object
	scm *azurescm.Azure
	// SourceBranch specifies the pullrequest source branch.
	SourceBranch string `yaml:",inline,omitempty"`
	// TargetBranch specifies the pullrequest target branch
	TargetBranch string `yaml:",inline,omitempty"`
}

// New returns a new valid Azure object.
func New(spec interface{}, scm *azurescm.Azure) (Azure, error) {

	var clientSpec client.Spec
	var s Spec

	// mapstructure.Decode cannot handle embedded fields
	// hence we decode it in two steps
	err := mapstructure.Decode(spec, &clientSpec)
	if err != nil {
		return Azure{}, err
	}

	err = mapstructure.Decode(spec, &s)
	if err != nil {
		return Azure{}, nil
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

		if len(clientSpec.Owner) == 0 && len(scm.Spec.Owner) > 0 {
			clientSpec.Owner = scm.Spec.Owner
		}

		if len(clientSpec.Project) == 0 && len(scm.Spec.Project) > 0 {
			clientSpec.Project = scm.Spec.Project
		}
	}

	// Sanitize modifies the clientSpec so it must be done once initialization is completed
	err = clientSpec.Sanitize()
	if err != nil {
		return Azure{}, err
	}

	c, err := client.New(clientSpec)

	if err != nil {
		return Azure{}, err
	}

	a := Azure{
		spec:   s,
		client: c,
		scm:    scm,
	}

	a.inheritFromScm()

	if err != nil {
		return Azure{}, nil
	}

	return a, nil
}

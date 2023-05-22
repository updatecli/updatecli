package mergerequest

import (
	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitlab/client"

	gitlabscm "github.com/updatecli/updatecli/pkg/plugins/scms/gitlab"
)

// Gitlab contains information to interact with GitLab api
type Gitlab struct {
	// spec contains inputs coming from updatecli configuration
	spec Spec
	// client handle the api authentication
	client client.Client
	// scm allows to interact with a scm object
	scm *gitlabscm.Gitlab
	// SourceBranch specifies the pullrequest source branch.
	SourceBranch string `yaml:",inline,omitempty"`
	// TargetBranch specifies the pullrequest target branch
	TargetBranch string `yaml:",inline,omitempty"`
	// Owner specifies repository owner
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// Repository specifies the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
}

// New returns a new valid GitLab object.
func New(spec interface{}, scm *gitlabscm.Gitlab) (Gitlab, error) {

	var clientSpec client.Spec
	var s Spec

	// mapstructure.Decode cannot handle embedded fields
	// hence we decode it in two steps
	err := mapstructure.Decode(spec, &clientSpec)
	if err != nil {
		return Gitlab{}, err
	}

	err = mapstructure.Decode(spec, &s)
	if err != nil {
		return Gitlab{}, nil
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

	c, err := client.New(clientSpec)

	if err != nil {
		return Gitlab{}, err
	}

	g := Gitlab{
		spec:   s,
		client: c,
		scm:    scm,
	}

	g.inheritFromScm()

	if err != nil {
		return Gitlab{}, nil
	}

	return g, nil

}

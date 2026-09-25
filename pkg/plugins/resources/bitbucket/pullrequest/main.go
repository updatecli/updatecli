package pullrequest

import (
	"fmt"

	"github.com/drone/go-scm/scm"
	"github.com/go-viper/mapstructure/v2"
	"github.com/updatecli/updatecli/pkg/plugins/resources/bitbucket/client"
	"github.com/updatecli/updatecli/pkg/plugins/scms/bitbucket"
)

/*
"bitbucket/pullrequest" defines the specification for opening or updating a Bitbucket Cloud pull request.
*/
type Spec struct {
	client.Spec
	// "sourcebranch" defines the branch the pull request merges from.
	//
	// default:
	//   the working branch of the associated scm.
	//
	SourceBranch string `yaml:",inline,omitempty"`
	// "targetbranch" defines the branch the pull request merges into.
	//
	// default:
	//   the branch of the associated scm.
	//
	TargetBranch string `yaml:",inline,omitempty"`
	// "owner" defines the repository owner.
	//
	// default:
	//   the owner of the associated scm.
	//
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// "repository" defines the repository name.
	//
	// default:
	//   the repository of the associated scm.
	//
	Repository string `yaml:",omitempty" jsonschema:"required"`
	// "title" defines the pull request title.
	//
	// default:
	//   the action title.
	//
	Title string `yaml:",inline,omitempty"`
	// "body" defines a custom pull request body.
	//
	// default:
	//   a report of the changes made by Updatecli.
	//
	// remark:
	//   * when set, it replaces the generated report.
	//
	Body string `yaml:",inline,omitempty"`
}

// Bitbucket contains information to interact with Bitbucket Cloud API
type Bitbucket struct {
	// spec contains inputs coming from updatecli configuration
	spec Spec
	// client handle the api authentication
	client *scm.Client
	// scm allows to interact with a scm object
	scm *bitbucket.Bitbucket
	// SourceBranch specifies the pullrequest source branch.
	SourceBranch string `yaml:",inline,omitempty"`
	// TargetBranch specifies the pullrequest target branch
	TargetBranch string `yaml:",inline,omitempty"`
	// Owner specifies repository owner
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// Repository specifies the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
}

// New returns a new valid Bitbucket Server object.
func New(spec any, scm *bitbucket.Bitbucket) (Bitbucket, error) {
	var clientSpec client.Spec
	var s Spec

	// mapstructure.Decode cannot handle embedded fields
	// hence we decode it in two steps
	err := mapstructure.Decode(spec, &clientSpec)
	if err != nil {
		return Bitbucket{}, fmt.Errorf("error decoding Bitbucket client spec: %w", err)
	}

	err = mapstructure.Decode(spec, &s)
	if err != nil {
		return Bitbucket{}, fmt.Errorf("error decoding Bitbucket pullrequest spec: %w", err)
	}

	if scm != nil {

		if len(clientSpec.Token) == 0 && len(scm.Spec.Token) > 0 {
			clientSpec.Token = scm.Spec.Token
		}

		if len(clientSpec.Repository) == 0 && len(scm.Spec.Repository) > 0 {
			clientSpec.Repository = scm.Spec.Repository
		}

		if len(clientSpec.Owner) == 0 && len(scm.Spec.Owner) > 0 {
			clientSpec.Owner = scm.Spec.Owner
		}

		if len(clientSpec.Username) == 0 && len(scm.Spec.Username) > 0 {
			clientSpec.Username = scm.Spec.Username
		}

		if len(clientSpec.Password) == 0 && len(scm.Spec.Password) > 0 {
			clientSpec.Password = scm.Spec.Password
		}
	}

	c, err := client.New(clientSpec)
	if err != nil {
		return Bitbucket{}, err
	}

	b := Bitbucket{
		spec:   s,
		client: c,
		scm:    scm,
	}

	b.inheritFromScm()

	return b, nil
}

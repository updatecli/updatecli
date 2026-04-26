package pullrequest

import (
	"fmt"

	"github.com/go-viper/mapstructure/v2"
	azdoclient "github.com/updatecli/updatecli/pkg/plugins/resources/azuredevops/client"
	azdoscm "github.com/updatecli/updatecli/pkg/plugins/scms/azuredevops"
)

// AzureDevOps contains information to interact with Azure DevOps pull requests.
type AzureDevOps struct {
	spec Spec
	// client handles the API authentication and helpers.
	client azdoclient.Client
	// scm allows to interact with a scm object.
	scm *azdoscm.AzureDevOps
	// SourceBranch specifies the pull request source branch.
	SourceBranch string `yaml:",omitempty"`
	// TargetBranch specifies the pull request target branch.
	TargetBranch string `yaml:",omitempty"`
	// Project specifies the Azure DevOps project.
	Project string `yaml:",omitempty" jsonschema:"required"`
	// Repository specifies the Azure DevOps repository.
	Repository string `yaml:",omitempty" jsonschema:"required"`
}

// New returns a new valid Azure DevOps action object.
func New(spec interface{}, scm *azdoscm.AzureDevOps) (AzureDevOps, error) {
	var clientSpec azdoclient.Spec
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return AzureDevOps{}, fmt.Errorf("error decoding spec: %w", err)
	}

	err = mapstructure.Decode(spec, &clientSpec)
	if err != nil {
		return AzureDevOps{}, fmt.Errorf("error decoding client spec: %w", err)
	}

	if clientSpec.URL == "" {
		clientSpec.URL = s.URL
	}

	if clientSpec.Project == "" {
		clientSpec.Project = s.Project
	}

	if clientSpec.Repository == "" {
		clientSpec.Repository = s.Repository
	}

	if clientSpec.Username == "" {
		clientSpec.Username = s.Username
	}

	if clientSpec.Organization == "" {
		clientSpec.Organization = s.Organization
	}

	if clientSpec.Token == "" {
		clientSpec.Token = s.Token
	}

	if scm != nil {
		if clientSpec.Token == "" && scm.Spec.Token != "" {
			clientSpec.Token = scm.Spec.Token
		}

		if clientSpec.URL == "" && scm.Spec.URL != "" {
			clientSpec.URL = scm.Spec.URL
		}

		if clientSpec.Username == "" && scm.Spec.Username != "" {
			clientSpec.Username = scm.Spec.Username
		}

		if clientSpec.Project == "" && scm.Spec.Project != "" {
			clientSpec.Project = scm.Spec.Project
		}

		if clientSpec.Repository == "" && scm.Spec.Repository != "" {
			clientSpec.Repository = scm.Spec.Repository
		}

		if clientSpec.Organization == "" && scm.Spec.Organization != "" {
			clientSpec.Organization = scm.Spec.Organization
		}
	}

	c, err := azdoclient.New(clientSpec)
	if err != nil {
		return AzureDevOps{}, err
	}

	s.Spec = c.Spec

	a := AzureDevOps{
		spec:   s,
		client: c,
		scm:    scm,
	}

	a.inheritFromScm()

	return a, nil
}

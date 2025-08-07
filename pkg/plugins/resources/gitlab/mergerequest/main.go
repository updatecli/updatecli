package mergerequest

import (
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitlab/client"

	gitlabscm "github.com/updatecli/updatecli/pkg/plugins/scms/gitlab"
	gitlabapi "gitlab.com/gitlab-org/api/client-go"
)

// Gitlab contains information to interact with GitLab api
type Gitlab struct {
	// spec contains inputs coming from updatecli configuration
	spec Spec
	// client handle the api authentication
	client client.Client
	// scm allows to interact with a scm object
	scm *gitlabscm.Gitlab
	// api allows to interact with Gitlab via API
	api *gitlabapi.Client
	// SourceBranch specifies the pullrequest source branch.
	SourceBranch string `yaml:",inline,omitempty"`
	// TargetBranch specifies the pullrequest target branch
	TargetBranch string `yaml:",inline,omitempty"`
	// Owner specifies repository owner
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// Repository specifies the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
}

func getGitlabClient(spec client.Spec) (*gitlabapi.Client, error) {
	tokenType := strings.ToLower(spec.TokenType)

	var opt gitlabapi.ClientOptionFunc
	if len(spec.URL) > 0 {
		opt = gitlabapi.WithBaseURL(spec.URL)
	}

	switch tokenType {
	case "bearer":
		return gitlabapi.NewOAuthClient(spec.Token, opt)
	case "private", "":
		return gitlabapi.NewClient(spec.Token, opt)
	default:
		return nil, fmt.Errorf("error: unknown tokenType '%s'", tokenType)
	}
}

// New returns a new valid GitLab object.
func New(spec interface{}, scm *gitlabscm.Gitlab) (Gitlab, error) {

	var clientSpec client.Spec
	var s Spec

	// mapstructure.Decode cannot handle embedded fields
	// hence we decode it in two steps
	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Gitlab{}, fmt.Errorf("error decoding spec: %w", err)
	}

	err = mapstructure.Decode(spec, &clientSpec)
	if err != nil {
		return Gitlab{}, fmt.Errorf("error decoding client spec: %w", err)
	}

	s.Spec = clientSpec

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

		if len(clientSpec.TokenType) == 0 && len(scm.Spec.TokenType) > 0 {
			clientSpec.TokenType = scm.Spec.TokenType
		}
	}

	c, err := client.New(clientSpec)

	if err != nil {
		return Gitlab{}, err
	}

	api, err := getGitlabClient(clientSpec)
	if err != nil {
		return Gitlab{}, err
	}

	g := Gitlab{
		spec:   s,
		client: c,
		scm:    scm,
		api:    api,
	}

	g.inheritFromScm()

	return g, nil

}

func (g *Gitlab) getPID() string {
	return strings.Join([]string{
		g.Owner,
		g.Repository}, "/")
}

package client

import (
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/credential"
)

// defaultTokenUsername is the git username used with a token when no username is set,
// as git authentication requires a non-empty username.
const defaultTokenUsername string = "oauth2"

// Spec defines the settings used to connect to a Gitea instance.
// It is shared by every "gitea" resource and by the "gitea/pullrequest" action.
type Spec struct {
	// "url" defines the Gitea url to interact with.
	//
	// remark:
	//   * "https://" is added when the url has no "https://" or "http://" prefix.
	//   * in a "gitea/pullrequest" action, the value is inherited from the scm when unset.
	//
	// example:
	//   * url: gitea.com
	//   * url: https://gitea.example.com
	//
	URL string `yaml:",omitempty" jsonschema:"required"`
	// "username" defines the username used to authenticate with the Gitea API.
	//
	// remark:
	//   * in a "gitea/pullrequest" action, the value is inherited from the scm when unset.
	//   * when unset, the environment variable "UPDATECLI_GITEA_USERNAME" is used.
	//   * when still unset and a token is available, "oauth2" is used.
	//
	Username string `yaml:",omitempty"`
	// "token" defines the credential used to authenticate with the Gitea API.
	//
	// remark:
	//   * a token is sensitive information. Do not set it directly in the manifest.
	//     Use an environment variable or a SOPS file instead.
	//   * `{{ requiredEnv "MY_TOKEN" }}` retrieves the token from the environment variable "MY_TOKEN".
	//   * `{{ .gitea.token }}` retrieves the token from a SOPS file.
	//     See https://github.com/getsops/sops for more information about SOPS files.
	//   * in a "gitea/pullrequest" action, the value is inherited from the scm when unset.
	//   * when unset, the environment variable "UPDATECLI_GITEA_TOKEN" is used.
	//     It applies to every Gitea server used by a pipeline.
	//
	Token string `yaml:",omitempty"`
}

// Validate validates that a spec contains good content
func (s Spec) Validate() error {

	if len(s.URL) == 0 {
		log.Errorf("missing %q parameter", "url")
		return fmt.Errorf("wrong configuration")
	}

	return nil
}

// Sanitize parse and update if needed a spec content
func (s *Spec) Sanitize() error {

	err := s.Validate()
	if err != nil {
		return err
	}

	if !strings.HasPrefix(s.URL, "https://") && !strings.HasPrefix(s.URL, "http://") {
		s.URL = "https://" + s.URL
	}

	username, token := credential.GetCredentialsFromEnv()

	if s.Username == "" {
		s.Username = username
	}

	if s.Token == "" {
		s.Token = token
	}

	if s.Username == "" && s.Token != "" {
		s.Username = defaultTokenUsername
	}

	return nil
}

package client

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

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
	//
	Username string `yaml:",omitempty"`
	// "token" defines the credential used to authenticate with the Gitea API.
	//
	// remark:
	//   * a token is sensitive information. Do not set it directly in the manifest.
	//     Use an environment variable or a SOPS file instead.
	//   * `{{ requiredEnv "GITEA_TOKEN" }}` retrieves the token from the environment variable "GITEA_TOKEN".
	//   * `{{ .gitea.token }}` retrieves the token from a SOPS file.
	//     See https://github.com/getsops/sops for more information about SOPS files.
	//   * in a "gitea/pullrequest" action, the value is inherited from the scm when unset.
	//
	Token string `yaml:",omitempty"`
}

// Validate validates that a spec contains good content
func (s Spec) Validate() error {

	if len(s.URL) == 0 {
		logrus.Errorf("missing %q parameter", "url")
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

	return nil
}

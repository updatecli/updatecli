package client

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// Spec defines a specification for a "gitea" resource
// parsed from an updatecli manifest file
type Spec struct {
	//  "url" defines the Gitea url to interact with
	URL string `yaml:",omitempty" jsonschema:"required"`
	//  "username" defines the username used to authenticate with Gitea API
	Username string `yaml:",omitempty"`
	//  "token" specifies the credential used to authenticate with Gitea API
	//
	//  remark:
	//    A token is a sensitive information, it's recommended to not set this value directly in the configuration file
	//    but to use an environment variable or a SOPS file.
	//
	//    The value can be set to `{{ requiredEnv "GITEA_TOKEN"}}` to retrieve the token from the environment variable `GITHUB_TOKEN`
	//	  or `{{ .gitea.token }}` to retrieve the token from a SOPS file.
	//
	//	  For more information, about a SOPS file, please refer to the following documentation:
	//    https://github.com/getsops/sops
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

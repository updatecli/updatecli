package client

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/drone/go-scm/scm"
	"github.com/drone/go-scm/scm/driver/stash"
	"github.com/drone/go-scm/scm/transport"
	"github.com/drone/go-scm/scm/transport/oauth2"
	"github.com/sirupsen/logrus"
)

// Spec defines a specification for a Bitbucket Server resource
// parsed from an updatecli manifest file
type Spec struct {
	// "url" specifies the default stash url in case of Bitbucket Server
	URL string `yaml:",omitempty" jsonschema:"required"`
	// "username" specifies the username used to authenticate with Bitbucket Server API
	Username string `yaml:",omitempty"`
	//  "token" specifies the credential used to authenticate with Bitbucket Server API
	//
	//  remark:
	//    A token is a sensitive information, it's recommended to not set this value directly in the configuration file
	//    but to use an environment variable or a SOPS file.
	//
	//    The value can be set to `{{ requiredEnv "BITBUCKET_TOKEN"}}` to retrieve the token from the environment variable `BITBUCKET_TOKEN`
	//	  or `{{ .bitbucket.token }}` to retrieve the token from a SOPS file.
	//
	//	  For more information, about a SOPS file, please refer to the following documentation:
	//    https://github.com/getsops/sops
	Token string `yaml:",omitempty"`
	//  "password" specifies the credential used to authenticate with Bitbucket Server API, it must be combined with "username"
	//
	//  remark:
	//    A token is a sensitive information, it's recommended to not set this value directly in the configuration file
	//    but to use an environment variable or a SOPS file.
	//
	//    The value can be set to `{{ requiredEnv "BITBUCKET_TOKEN"}}` to retrieve the token from the environment variable `BITBUCKET_TOKEN`
	//	  or `{{ .bitbucket.token }}` to retrieve the token from a SOPS file.
	//
	//	  For more information, about a SOPS file, please refer to the following documentation:
	//    https://github.com/getsops/sops
	Password string `yaml:",omitempty"`
	// "owner" defines repository owner
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// "repository" defines the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
}

type Client *scm.Client

func New(s Spec) (Client, error) {
	client, err := stash.New(s.URL)
	if err != nil {
		return nil, err
	}

	if len(s.Token) >= 0 {
		if len(s.Username) >= 0 {
			client.Client = &http.Client{
				Transport: &transport.BasicAuth{
					Username: s.Username,
					Password: s.Token,
				},
			}
		} else {
			client.Client = &http.Client{
				Transport: &oauth2.Transport{
					Source: oauth2.StaticTokenSource(
						&scm.Token{
							Token: s.Token,
						},
					),
				},
			}
		}
	}

	return client, nil
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

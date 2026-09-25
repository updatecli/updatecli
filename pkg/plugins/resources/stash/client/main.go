package client

import (
	"fmt"
	"strings"

	"github.com/drone/go-scm/scm"
	"github.com/drone/go-scm/scm/driver/stash"
	"github.com/drone/go-scm/scm/transport"
	"github.com/drone/go-scm/scm/transport/oauth2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
)

// Spec defines the settings used to connect to a Bitbucket Server API.
type Spec struct {
	// "url" defines the Bitbucket Server url.
	//
	// remark:
	//   * "url" is required.
	//   * "https://" is added when the url has no scheme.
	//
	// example:
	//   * url: bitbucket.example.com
	//
	URL string `yaml:",omitempty" jsonschema:"required"`
	// "username" defines the username used to authenticate with the Bitbucket Server API.
	//
	// remark:
	//   * when set, "token" is sent as the password of a basic authentication.
	//
	Username string `yaml:",omitempty"`
	// "token" defines the credential used to authenticate with the Bitbucket Server API.
	//
	// remark:
	//   * without "username", the token is sent as a bearer token.
	//   * a token is sensitive information. Do not set it directly in the manifest,
	//     use an environment variable or a SOPS file instead.
	//   * the value can be set to `{{ requiredEnv "BITBUCKET_TOKEN"}}` to read the token
	//     from the environment variable `BITBUCKET_TOKEN`, or to `{{ .bitbucket.token }}`
	//     to read it from a SOPS file.
	//   * more information about SOPS on https://github.com/getsops/sops
	//
	Token string `yaml:",omitempty"`
	// "password" defines the credential used to authenticate with the Bitbucket Server API.
	//
	// remark:
	//   * it must be combined with "username".
	//   * the Bitbucket Server client does not read "password" yet: set the credential
	//     in "token" together with "username" for a basic authentication.
	//   * a password is sensitive information. Do not set it directly in the manifest,
	//     use an environment variable or a SOPS file instead.
	//   * the value can be set to `{{ requiredEnv "BITBUCKET_PASSWORD"}}` to read it
	//     from the environment variable `BITBUCKET_PASSWORD`, or to `{{ .bitbucket.password }}`
	//     to read it from a SOPS file.
	//   * more information about SOPS on https://github.com/getsops/sops
	//
	Password string `yaml:",omitempty"`
	// "owner" defines the repository owner.
	//
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// "repository" defines the repository name.
	//
	Repository string `yaml:",omitempty" jsonschema:"required"`
}

type Client *scm.Client

func New(s Spec) (Client, error) {
	client, err := stash.New(s.URL)
	if err != nil {
		return nil, err
	}

	client.Client = httpclient.NewRetryClient()

	if s.Token != "" {
		if s.Username != "" {
			client.Client.Transport = &transport.BasicAuth{
				Username: s.Username,
				Password: s.Token,
				Base:     client.Client.Transport,
			}
		} else {
			client.Client.Transport = &oauth2.Transport{
				Source: oauth2.StaticTokenSource(
					&scm.Token{
						Token: s.Token,
					},
				),
				Base: client.Client.Transport,
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

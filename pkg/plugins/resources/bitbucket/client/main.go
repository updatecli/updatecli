package client

import (
	"github.com/drone/go-scm/scm"
	"github.com/drone/go-scm/scm/driver/bitbucket"
	"github.com/drone/go-scm/scm/transport"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
)

// Spec defines the settings used to connect to a Bitbucket Cloud repository.
type Spec struct {
	// "username" defines the username used to authenticate with the Bitbucket Cloud API.
	//
	// remark:
	//   * it is combined with "password".
	//
	Username string `yaml:",omitempty"`
	// "token" defines the credential used to authenticate with the Bitbucket Cloud API.
	//
	// remark:
	//   * it is a repository or project access token with the "pullrequest:write" scope.
	//   * set either "token" or "password". When both are set, "token" is used.
	//   * a token is sensitive information. Do not write it directly in the manifest,
	//     use an environment variable or a SOPS file instead.
	//   * `{{ requiredEnv "BITBUCKET_TOKEN" }}` reads the token from the environment variable "BITBUCKET_TOKEN".
	//   * `{{ .bitbucket.token }}` reads the token from a SOPS file,
	//     see https://github.com/getsops/sops
	//
	Token string `yaml:",omitempty"`
	// "password" defines the credential used to authenticate with the Bitbucket Cloud API.
	//
	// remark:
	//   * it must be combined with "username".
	//   * it should be an app password with the "pullrequest:write" scope.
	//   * set either "token" or "password". When both are set, "token" is used.
	//   * a password is sensitive information. Do not write it directly in the manifest,
	//     use an environment variable or a SOPS file instead.
	//   * `{{ requiredEnv "BITBUCKET_PASSWORD" }}` reads the password from the environment variable "BITBUCKET_PASSWORD".
	//   * `{{ .bitbucket.password }}` reads the password from a SOPS file,
	//     see https://github.com/getsops/sops
	//
	Password string `yaml:",omitempty"`
	// "owner" defines the repository owner.
	//
	// example:
	//   * owner: updatecli
	//
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// "repository" defines the repository name.
	//
	// example:
	//   * repository: website
	//
	Repository string `yaml:",omitempty" jsonschema:"required"`
}

func New(s Spec) (*scm.Client, error) {
	client := bitbucket.NewDefault()

	client.Client = httpclient.NewRetryClient()

	if (len(s.Username) > 0) && (len(s.Password) > 0) {
		client.Client.Transport = &transport.BasicAuth{
			Username: s.Username,
			Password: s.Password,
			Base:     client.Client.Transport,
		}
	}

	if len(s.Token) > 0 {
		client.Client.Transport = &transport.BearerToken{
			Token: s.Token,
			Base:  client.Client.Transport,
		}
	}

	return client, nil
}

func URL() string {
	return "https://bitbucket.org"
}

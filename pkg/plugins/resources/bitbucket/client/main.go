package client

import (
	"net/http"

	"github.com/drone/go-scm/scm"
	"github.com/drone/go-scm/scm/driver/bitbucket"
	"github.com/drone/go-scm/scm/transport"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
)

// Spec defines a specification for a Bitbucket Cloud resource
// parsed from an updatecli manifest file
type Spec struct {
	// "username" specifies the username used to authenticate with Bitbucket Cloud API
	Username string `yaml:",omitempty"`
	//  "token" specifies the credential used to authenticate with Bitbucket Cloud API
	//
	//  The "token" is a repository or project access token with "pullrequest:write" scope.
	//
	//  "token" and "password" are mutually exclusive
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
	//  "password" specifies the credential used to authenticate with Bitbucket Cloud API, it must be combined with "username"
	//
	//  The "password" should be app password with "pullrequest:write" scope.
	//
	//  "token" and "password" are mutually exclusive
	//
	//  remark:
	//    A password is a sensitive information, it's recommended to not set this value directly in the configuration file
	//    but to use an environment variable or a SOPS file.
	//
	//    The value can be set to `{{ requiredEnv "BITBUCKET_PASSWORD"}}` to retrieve the token from the environment variable `BITBUCKET_PASSWORD`
	//	  or `{{ .bitbucket.password }}` to retrieve the token from a SOPS file.
	//
	//	  For more information, about a SOPS file, please refer to the following documentation:
	//    https://github.com/getsops/sops
	Password string `yaml:",omitempty"`
	// "owner" defines repository owner
	Owner string `yaml:",omitempty" jsonschema:"required"`
	// "repository" defines the name of a repository for a specific owner
	Repository string `yaml:",omitempty" jsonschema:"required"`
}

func New(s Spec) (*scm.Client, error) {
	client := bitbucket.NewDefault()

	client.Client = httpclient.NewRetryClient().(*http.Client)

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

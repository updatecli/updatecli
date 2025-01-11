package client

import (
	"net/http"
	"strings"

	"github.com/drone/go-scm/scm"
	"github.com/drone/go-scm/scm/driver/gitlab"
	"github.com/drone/go-scm/scm/transport"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
)

const (

	// GITLABDOMAIN defines the default gitlab url
	GITLABDOMAIN string = "gitlab.com"
)

type Client *scm.Client

func New(s Spec) (Client, error) {

	url := EnsureValidURL(s.URL)

	client, err := gitlab.New(url)

	if err != nil {
		return nil, err
	}

	client.Client = httpclient.NewRetryClient().(*http.Client)

	if len(s.Token) >= 0 {
		// provide a custom http.Client with a transport
		// that injects the private GitLab token through
		// the PRIVATE_TOKEN header variable.
		client.Client.Transport = &transport.PrivateToken{
			Token: s.Token,
			Base:  client.Client.Transport,
		}
	}

	return client, nil

}

func EnsureValidURL(u string) string {
	if u == "" {
		u = GITLABDOMAIN
	}

	if !strings.HasPrefix(u, "https://") && !strings.HasPrefix(u, "http://") {
		u = "https://" + u
	}

	return u
}

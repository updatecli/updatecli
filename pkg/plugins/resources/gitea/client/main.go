package client

import (
	"net/http"

	"github.com/drone/go-scm/scm"
	"github.com/drone/go-scm/scm/driver/gitea"
	"github.com/drone/go-scm/scm/transport/oauth2"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
)

type Client *scm.Client

func New(s Spec) (Client, error) {

	client, err := gitea.New(s.URL)

	if err != nil {
		return nil, err
	}

	client.Client = httpclient.NewRetryClient().(*http.Client)

	if len(s.Token) >= 0 {
		client.Client.Transport = &oauth2.Transport{
			Source: oauth2.StaticTokenSource(
				&scm.Token{
					Token: s.Token,
				},
			),
			Base: client.Client.Transport,
		}
	}

	return client, nil
}

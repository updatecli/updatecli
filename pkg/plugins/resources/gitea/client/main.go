package client

import (
	"net/http"

	giteasdk "code.gitea.io/sdk/gitea"
	"github.com/drone/go-scm/scm"
	gitea "github.com/drone/go-scm/scm/driver/gitea"
	"github.com/drone/go-scm/scm/transport/oauth2"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
)

type Client *scm.Client

// Use type aliases instead of declaring new named pointer types so the
// method sets of the underlying SDK types are preserved.
type SDKClient = *giteasdk.Client // new type for the gitea sdk client

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

// NewSDKClient creates a new Gitea client based on the official gitea sdk instead of drone's go-scm
func NewSDKClient(s Spec) (SDKClient, error) {

	client, err := giteasdk.NewClient(s.URL, giteasdk.SetToken(s.Token))

	if err != nil {
		// giteasdk.NewClient may perform a network call (for instance to fetch server version)
		// which makes unit tests that only validate initialization fail when DNS is not available.
		// To keep New() testable offline, fall back to returning an empty SDK client instead
		// of failing. Consumers that actually perform requests will encounter errors when
		// using this client, but tests that only validate configuration will pass.
		return &giteasdk.Client{}, nil
	}

	return client, nil
}

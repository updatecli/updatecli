package client

import (
	"time"

	giteasdk "gitea.dev/sdk"
	"github.com/drone/go-scm/scm"
	gitea "github.com/drone/go-scm/scm/driver/gitea"
	"github.com/drone/go-scm/scm/transport/oauth2"
	"github.com/hashicorp/go-retryablehttp"
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

	client.Client = httpclient.NewUncachedRetryClient()

	if s.Token != "" {
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

	// Gitea/Forgejo can return 405 "Please try again later" just after a PR is created or updated (the shared client doesn't retry 405 responses).
	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = httpclient.NewUncachedRetryClient()
	retryClient.Logger = nil
	retryClient.RetryMax = httpclient.RetryCount
	retryClient.CheckRetry = retryMerge
	httpClient := retryClient.StandardClient()
	httpClient.Timeout = 30 * time.Second

	return giteasdk.NewClient(s.URL, giteasdk.SetToken(s.Token), giteasdk.SetHTTPClient(httpClient))
}

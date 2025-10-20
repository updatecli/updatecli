package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github/app"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github/token"
	"golang.org/x/oauth2"
)

// Client must be implemented by any GitHub query client (v4 API)
type Client interface {
	Query(ctx context.Context, q interface{}, variables map[string]interface{}) error
	Mutate(ctx context.Context, m interface{}, input githubv4.Input, variables map[string]interface{}) error
}

// ClientConfig defines the configuration required to create a GitHub client instance.
type ClientConfig struct {
	Username    string
	Client      Client
	TokenSource oauth2.TokenSource
	URL         string
}

var (
	// MaxRetry defines the maximum number of retries when rate limit is exceeded
	MaxRetry = 3
)

// New creates a new GitHub client instance based on the provided configuration.
func New(configUsername, configToken string, configApp *app.Spec, configURL string) (client *ClientConfig, err error) {

	URL := "github.com"
	if configURL != "" {
		URL = strings.TrimSpace(configURL)
	}

	if !strings.HasPrefix(URL, "https://") && !strings.HasPrefix(URL, "http://") {
		URL = "https://" + URL
	}

	// We first try to get a token source from the environment variable
	username, tokenSource, err := token.GetTokenSourceFromEnv()
	if err != nil {
		logrus.Debugf("no GitHub token found in environment variables: %s", err)
	}

	// If no token source could be found in the environment variable
	// we try to get it from the configuration
	if tokenSource == nil {
		username, tokenSource, err = token.GetTokenSourceFromConfig(configUsername, configToken, configApp)
		if err != nil {
			return nil, fmt.Errorf("retrieving token source from configuration: %w", err)
		}
	}

	if tokenSource == nil {
		username, tokenSource = token.GetFallbackTokenSourceFromEnv()
	}

	// If the tokenSource is still nil at this point
	// it means that no valid token source could be found.
	// We log a debug message and return an error.
	if tokenSource == nil {
		logrus.Debugf(`GitHub token is not set, please refer to the documentation for more information:
	->  https://www.updatecli.io/docs/plugins/scm/github/
`)
		return nil, errors.New("github token is not set")
	}

	tokenSource = oauth2.ReuseTokenSource(nil, tokenSource)

	clientContext := context.WithValue(
		context.Background(),
		oauth2.HTTPClient,
		httpclient.NewRetryClient().(*http.Client))

	httpClient := oauth2.NewClient(clientContext, tokenSource)

	var newClient Client

	if strings.HasSuffix(URL, "github.com") {
		newClient = githubv4.NewClient(httpClient)
	} else {
		// For GH enterprise the GraphQL API path is /api/graphql
		// Cf https://docs.github.com/en/enterprise-cloud@latest/graphql/guides/managing-enterprise-accounts#3-setting-up-insomnia-to-use-the-github-graphql-api-with-enterprise-accounts
		graphqlURL, err := url.JoinPath(URL, "/api/graphql")
		if err != nil {
			return nil, fmt.Errorf("parsing GitHub Enterprise GraphQL URL: %w", err)
		}
		newClient = githubv4.NewEnterpriseClient(graphqlURL, httpClient)
	}

	return &ClientConfig{
		Username:    username,
		Client:      newClient,
		TokenSource: tokenSource,
		URL:         URL,
	}, nil
}

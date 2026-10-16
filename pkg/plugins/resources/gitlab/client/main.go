package client

import (
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/httpclient"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"golang.org/x/time/rate"
)

const (

	// GITLABDOMAIN defines the default gitlab url
	GITLABDOMAIN string = "gitlab.com"
)

type Client *gitlab.Client

func New(s Spec) (Client, error) {
	var client *gitlab.Client
	var err error

	url := EnsureValidURL(s.URL)

	/*
		The client's default limiter derives its rate from the first response's
		Ratelimit-Limit header, assuming a per minute value. gitlab.com also answers
		unauthenticated calls with throttles that don't follow that unit, such as
		unauthenticated_traffic_per_ip reporting a limit of 1, which throttles the
		client to one request every ninety seconds. Rate limiting is left to the
		server instead, as the client still retries 429 responses, waiting until
		the Ratelimit-Reset header when the server sends it.

		Retries are left to the GitLab client alone, so the HTTP client it wraps
		doesn't retry. Unlike the updatecli retry client, it doesn't retry a write
		after a transport error that may have reached the server, which could
		otherwise create a merge request, branch, tag or release twice. It still
		retries any request, writes included, on 5xx responses other than 501.
		The client also writes, so it bypasses the HTTP cache.
	*/
	client, err = gitlab.NewClient(
		s.Token,
		gitlab.WithBaseURL(url),
		gitlab.WithHTTPClient(httpclient.NewUncachedClient()),
		gitlab.WithUserAgent(httpclient.UserAgent()),
		gitlab.WithCustomLimiter(rate.NewLimiter(rate.Inf, 0)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
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

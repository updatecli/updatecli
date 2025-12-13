package client

import (
	"fmt"
	"strings"

	gitlab "gitlab.com/gitlab-org/api/client-go"
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

	client, err = gitlab.NewClient(s.Token, gitlab.WithBaseURL(url))
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

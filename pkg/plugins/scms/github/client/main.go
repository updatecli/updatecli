package client

import (
	"context"

	"github.com/shurcooL/githubv4"
)

// Client must be implemented by any GitHub query client (v4 API)
type Client interface {
	Query(ctx context.Context, q interface{}, variables map[string]interface{}) error
	Mutate(ctx context.Context, m interface{}, input githubv4.Input, variables map[string]interface{}) error
}

var (
	MaxRetry = 3
)

package githubsearch

import (
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github/client"
)

const (
	DefaultRepositoryLimit = 10
	ErrSearchQueryEmpty    = "GitHub search query is required"
	Kind                   = "githubsearch"
)

type GitHubSearch struct {
	spec   Spec
	limit  int
	branch string
	search string
	client client.Client
}

func New(s interface{}) (*GitHubSearch, error) {
	var spec Spec

	// mapstructure.Decode cannot handle embedded fields
	// hence we decode it in two steps
	err := mapstructure.Decode(s, &spec)
	if err != nil {
		return nil, err
	}

	if err := spec.Validate(); err != nil {
		return nil, err
	}

	limit := DefaultRepositoryLimit
	if spec.Limit != nil {
		limit = *spec.Limit
	}

	branch := "^main$"
	if spec.Branch != "" {
		branch = spec.Branch
	}

	search := strings.TrimSpace(spec.Search)

	clientConfig, err := client.New(spec.Username, spec.Token, spec.App, spec.URL)
	if err != nil {
		return nil, fmt.Errorf("creating GitHub client: %w", err)
	}

	return &GitHubSearch{
		spec:   spec,
		limit:  limit,
		branch: branch,
		search: search,
		client: clientConfig.Client,
	}, nil
}

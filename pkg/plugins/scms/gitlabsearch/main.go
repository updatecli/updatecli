package gitlabsearch

import (
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitlab/client"
)

const (
	DefaultRepositoryLimit = 10
	ErrGroupEmpty          = "GitLab group is required for gitlabsearch SCM"
	Kind                   = "gitlabsearch"
)

// GitLabSearch holds configuration for searching GitLab repositories within a group.
type GitLabSearch struct {
	spec             Spec
	limit            int
	branch           string
	group            string
	search           string
	includeSubgroups bool
	client           client.Client
}

// New creates a new GitLabSearch instance from the provided configuration.
func New(s interface{}) (*GitLabSearch, error) {
	var spec Spec
	var clientSpec client.Spec

	// mapstructure.Decode cannot handle embedded fields
	// hence we decode it in two steps
	err := mapstructure.Decode(s, &clientSpec)
	if err != nil {
		return nil, err
	}

	err = mapstructure.Decode(s, &spec)
	if err != nil {
		return nil, err
	}

	spec.Spec = clientSpec

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

	includeSubgroups := true
	if spec.IncludeSubgroups != nil {
		includeSubgroups = *spec.IncludeSubgroups
	}

	c, err := client.New(clientSpec)
	if err != nil {
		return nil, fmt.Errorf("creating GitLab client: %w", err)
	}

	return &GitLabSearch{
		spec:             spec,
		limit:            limit,
		branch:           branch,
		group:            strings.TrimSpace(spec.Group),
		search:           strings.TrimSpace(spec.Search),
		includeSubgroups: includeSubgroups,
		client:           c,
	}, nil
}

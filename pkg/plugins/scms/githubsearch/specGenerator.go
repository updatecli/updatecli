package githubsearch

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
)

var (
	// Manifests of a compose file or a policy usually share one search, which ran the
	// same queries once per manifest.
	queryResults     = map[string][]string{}
	queryResultsLock sync.Mutex
)

// cachedQuery runs query once per execution for the same credentials and key.
func (g GitHubSearch) cachedQuery(key string, query func() ([]string, error)) ([]string, error) {
	app := ""
	if g.spec.App != nil {
		app = fmt.Sprintf("%+v", *g.spec.App)
	}
	key = strings.Join([]string{g.spec.URL, g.spec.Username, g.spec.Token, app, key}, "\x00")

	queryResultsLock.Lock()
	defer queryResultsLock.Unlock()

	if result, ok := queryResults[key]; ok {
		return result, nil
	}
	result, err := query()
	if err != nil {
		return nil, err
	}
	queryResults[key] = result
	return result, nil
}

// ScmsGenerator generates GitHub SCM specs based on the search query and branch filter.
func (g GitHubSearch) ScmsGenerator(ctx context.Context) (results []github.Spec, err error) {

	results = make([]github.Spec, 0)

	repositories, err := g.cachedQuery("search\x00"+g.search, func() ([]string, error) {
		return github.SearchRepositories(g.client, g.search, 0, ctx)
	})
	if err != nil {
		return nil, fmt.Errorf("failed generating spec: %w", err)
	}

	for _, repo := range repositories {
		logrus.Debugf("Processing GitHub repository: %s", repo)

		repositoryParts := strings.Split(repo, "/")
		if len(repositoryParts) != 2 {
			return nil, fmt.Errorf("invalid repository format: %s", repo)
		}

		branches, err := g.cachedQuery("branches\x00"+repo, func() ([]string, error) {
			return github.ListBranches(g.client, repositoryParts[0], repositoryParts[1], 0, ctx)
		})

		for _, b := range branches {

			re := regexp.MustCompile(g.branch)
			if !re.MatchString(b) {
				continue
			}

			spec := github.Spec{
				App:                    g.spec.App,
				Branch:                 b,
				CommitMessage:          g.spec.CommitMessage,
				CommitUsingAPI:         g.spec.CommitUsingAPI,
				Directory:              g.spec.Directory,
				Depth:                  g.spec.Depth,
				Email:                  g.spec.Email,
				Force:                  g.spec.Force,
				GPG:                    g.spec.GPG,
				Owner:                  repositoryParts[0],
				Repository:             repositoryParts[1],
				Submodules:             g.spec.Submodules,
				Token:                  g.spec.Token,
				URL:                    g.spec.URL,
				Username:               g.spec.Username,
				User:                   g.spec.User,
				WorkingBranch:          g.spec.WorkingBranch,
				WorkingBranchPrefix:    g.spec.WorkingBranchPrefix,
				WorkingBranchSeparator: g.spec.WorkingBranchSeparator,
			}
			results = append(results, spec)

			if g.limit > 0 && len(results) >= g.limit {
				return results, nil
			}
		}

		if err != nil {
			return nil, fmt.Errorf("failed generating GitHub scm: %w", err)
		}
	}

	return results, nil
}

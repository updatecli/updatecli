package githubsearch

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
)

// ScmsGenerator generates GitHub SCM specs based on the search query and branch filter.
func (g GitHubSearch) ScmsGenerator(ctx context.Context) (results []github.Spec, err error) {

	results = make([]github.Spec, 0)

	repositories, err := github.SearchRepositories(g.client, g.search, 0, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed generating spec: %w", err)
	}

	for _, repo := range repositories {
		logrus.Debugf("Processing GitHub repository: %s", repo)

		repositoryParts := strings.Split(repo, "/")
		if len(repositoryParts) != 2 {
			return nil, fmt.Errorf("invalid repository format: %s", repo)
		}

		branches, err := github.ListBranches(g.client, repositoryParts[0], repositoryParts[1], 0, ctx)

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

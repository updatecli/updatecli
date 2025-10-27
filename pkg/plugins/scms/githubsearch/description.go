package githubsearch

import (
	"fmt"
)

// Summary returns a brief description of the GitHub search SCM configuration
func (g *GitHubSearch) Summary() string {
	return fmt.Sprintf("GitHub Advanced Search:\n\tQuery: %q\n\tBranch: %q\n\tLimit: %d", g.search, g.branch, g.limit)
}

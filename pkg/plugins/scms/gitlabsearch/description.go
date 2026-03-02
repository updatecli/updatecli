package gitlabsearch

import (
	"fmt"
)

// Summary returns a brief description of the GitLab search SCM configuration.
func (g *GitLabSearch) Summary() string {
	return fmt.Sprintf("GitLab Group Search:\n\tGroup: %q\n\tBranch: %q\n\tLimit: %d", g.group, g.branch, g.limit)
}

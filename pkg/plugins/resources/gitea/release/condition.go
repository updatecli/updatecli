package release

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
)

func (g *Gitea) Condition(ctx context.Context, source string, scm scm.ScmHandler) (pass bool, message string, err error) {

	if scm != nil {
		logrus.Warningf("Condition not supported for the plugin Gitea Release")
	}

	// A condition checks whether a specific release exists, so the age filter,
	// which only narrows down which release to pick, doesn't apply here.
	releases, err := g.SearchReleases(ctx, age.Spec{})
	if err != nil {
		return false, "", fmt.Errorf("looking for Gitea release: %w", err)
	}

	release := source
	if g.spec.Tag != "" {
		release = g.spec.Tag
	}

	if len(releases) == 0 {
		return false, "", fmt.Errorf("no Gitea release found")
	}

	for _, r := range releases {
		if r == release {
			return true, fmt.Sprintf("Gitea release %q found", release), nil
		}
	}

	return false, fmt.Sprintf("no Gitea Release tag found matching pattern %q", g.versionFilter.Pattern), nil
}

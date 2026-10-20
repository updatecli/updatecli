package tag

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

func (g *Gitea) Condition(_ context.Context, source string, scm scm.ScmHandler, pathResolver pathresolver.Resolver) (pass bool, message string, err error) {
	if scm != nil {
		log.Warningf("Condition not supported for the plugin Gitea tag")
	}

	tag := source
	if g.spec.Tag != "" {
		tag = g.spec.Tag
	}

	tags, err := g.SearchTags()
	if err != nil {
		return false, "", fmt.Errorf("looking for Gitea tag: %w", err)
	}

	if len(tags) == 0 {
		return false, "", fmt.Errorf("no Gitea Tags found")
	}

	for _, t := range tags {
		if t == tag {
			return true, fmt.Sprintf("Gitea tag %q found", t), nil
		}
	}

	return false, fmt.Sprintf("no Gitea tag found matching %q", g.spec.Tag), nil
}

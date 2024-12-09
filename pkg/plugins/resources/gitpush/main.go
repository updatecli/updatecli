package gitpush

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/reports"
)

// GitPush defines a GitPush action
type GitPush struct {
	scmHandler scm.ScmHandler
}

// NewAction Initializes a new GitPush action
func NewAction(s scm.Scm) (GitPush, error) {
	supportedScms := map[string]int{
		scm.GitIdentifier:    1,
		scm.GithubIdentifier: 1,
		scm.StashIdentifier:  1,
		scm.GitlabIdentifier: 1,
		scm.GiteaIdentifier:  1,
	}
	_, exists := supportedScms[s.Config.Kind]
	if !exists {
		return GitPush{}, fmt.Errorf("scm of kind %q is not compatible with action of kind %q",
			s.Config.Kind,
			"git")
	}
	return GitPush{
		scmHandler: s.Handler,
	}, nil
}

// CreateAction executes a "push" on the associated SCM
func (gp *GitPush) CreateAction(report *reports.Action, resetDescription bool) error {
	_, err := gp.scmHandler.Push()
	return err
}

func (gp *GitPush) CleanAction(report *reports.Action) error {
	return nil
}

func (gp *GitPush) CheckActionExist(report *reports.Action) error {
	return nil
}

package ci

import (
	"fmt"
	"os"
)

// GiteaActions detects Gitea Actions, which also exposes GitHub Actions variables.
type GiteaActions struct {
}

func (ga GiteaActions) Name() string {
	return "Gitea Actions workflow link"
}

// URL returns the link to the current run. Recent Gitea versions read the run
// path as a run ID first, so by_index=1 makes them read it as the run number.
func (ga GiteaActions) URL() string {
	return fmt.Sprintf("%s/%s/actions/runs/%s?by_index=1", os.Getenv("GITHUB_SERVER_URL"), os.Getenv("GITHUB_REPOSITORY"), os.Getenv("GITHUB_RUN_NUMBER"))
}

func (ga GiteaActions) IsDebug() bool {
	return (os.Getenv("ACTIONS_RUNNER_DEBUG") == True || os.Getenv("ACTIONS_STEP_DEBUG") == True)
}

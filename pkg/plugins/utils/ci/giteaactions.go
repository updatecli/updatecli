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

func (ga GiteaActions) URL() string {
	return fmt.Sprintf("%s/%s/actions/runs/%s", os.Getenv("GITHUB_SERVER_URL"), os.Getenv("GITHUB_REPOSITORY"), os.Getenv("GITHUB_RUN_NUMBER"))
}

func (ga GiteaActions) IsDebug() bool {
	return (os.Getenv("ACTIONS_RUNNER_DEBUG") == True || os.Getenv("ACTIONS_STEP_DEBUG") == True)
}

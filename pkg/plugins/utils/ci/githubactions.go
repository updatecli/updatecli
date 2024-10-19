package ci

import (
	"fmt"
	"os"
)

type GitHubActions struct {
}

func (gha GitHubActions) Name() string {
	return "GitHub Action workflow link"
}

func (gha GitHubActions) URL() string {
	return fmt.Sprintf(os.Getenv("GITHUB_SERVER_URL")+"/%s/actions/runs/%s", os.Getenv("GITHUB_REPOSITORY"), os.Getenv("GITHUB_RUN_ID"))
}

func (gha GitHubActions) IsDebug() bool {
	return (os.Getenv("ACTIONS_RUNNER_DEBUG") == True || os.Getenv("ACTIONS_STEP_DEBUG") == True)
}

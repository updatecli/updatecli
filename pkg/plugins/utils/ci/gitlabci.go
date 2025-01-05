package ci

import (
	"fmt"
	"os"
)

type GitLabCi struct {
}

func (g GitLabCi) Name() string {
	return "GitLab CI pipeline link"
}

func (g GitLabCi) URL() string {
	return fmt.Sprintf(os.Getenv("CI_SERVER_URL")+"/%s/-/jobs/%s", os.Getenv("CI_PROJECT_PATH"), os.Getenv("CI_JOB_ID"))
}

func (gha GitLabCi) IsDebug() bool {
	// See https://docs.gitlab.com/ee/ci/variables/index.html#enable-debug-logging
	return os.Getenv("CI_DEBUG_TRACE") == True
}

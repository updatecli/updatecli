package ci

import (
	"fmt"
	"os"
)

var True = "true"

// CIEngine is an interface which allows to detects based on environment variable if Updatecli is executed from a CI environment like Jenkins or GitLab CI
type CIEngine interface {
	URL() string
	Name() string
	IsDebug() bool
}

// New returns a newly initialized CIEngine or an error
func New() (ci CIEngine, err error) {
	if os.Getenv("JENKINS_URL") != "" {
		return Jenkins{}, nil
	}

	if os.Getenv("GITLAB_CI") != "" {
		return GitLabCi{}, nil
	}

	if os.Getenv("GITHUB_ACTION") != "" {
		return GitHubActions{}, nil
	}

	return nil, fmt.Errorf("Unknown CI Engine.")
}

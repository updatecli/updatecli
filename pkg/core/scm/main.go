package scm

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/plugins/git"
	"github.com/updatecli/updatecli/pkg/plugins/github"
)

// Scm is an interface that offers common functions for a source control manager like git or github
type Scm interface {
	Add(files []string) error
	Clone() (string, error)
	Checkout() error
	GetDirectory() (directory string)
	Init(source string, pipelineID string) error
	Push() error
	Commit(message string) error
	Clean() error
	PushTag(tag string) error
	GetChangedFiles(workingDir string) ([]string, error)
}

// Unmarshal parses a scm struct like git or github and returns a scm interface
func Unmarshal(scm map[string]interface{}) (Scm, PullRequest, error) {
	var s Scm
	var pr PullRequest
	if len(scm) != 1 {
		return nil, nil, fmt.Errorf("target scm: only one scm can be provided between git and github")
	}

	for key, value := range scm {
		switch key {
		case "github":

			githubSpec := github.Spec{}

			err := mapstructure.Decode(value, &githubSpec)
			if err != nil {
				return nil, nil, err
			}

			g, err := github.New(githubSpec)

			if err != nil {
				return nil, nil, err
			}

			s = &g
			pr = &g

		case "git":
			g := git.Git{}

			err := mapstructure.Decode(value, &g)
			if err != nil {
				return nil, nil, err
			}

			s = &g
		default:
			return nil, nil, fmt.Errorf("wrong scm type provided, accepted values [git,github]")
		}
	}
	return s, pr, nil
}

package scm

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/olblak/updateCli/pkg/plugins/git"
	"github.com/olblak/updateCli/pkg/plugins/github"
)

// Scm is an interface that offers common functions for a source control manager like git or github
type Scm interface {
	Add(files []string) error
	Clone() (string, error)
	Checkout() error
	GetDirectory() (directory string)
	Init(source string, name string) error
	Push() error
	Commit(message string) error
	Clean() error
}

// Unmarshal parses a scm struct like git or github and returns a scm interface
func Unmarshal(scm map[string]interface{}) (Scm, error) {
	var s Scm
	if len(scm) != 1 {
		return nil, fmt.Errorf("Target scm: Only one scm can be provided between git and github")
	}

	for key, value := range scm {
		switch key {
		case "github":

			var g github.Github

			err := mapstructure.Decode(value, &g)

			if err != nil {
				return nil, err
			}

			s = &g

		case "git":
			g := git.Git{}

			err := mapstructure.Decode(value, &g)

			if err != nil {
				return nil, err
			}

			s = &g

		default:
			return nil, fmt.Errorf("wrong scm type provided, accepted values [git,github]")

		}
	}
	return s, nil
}

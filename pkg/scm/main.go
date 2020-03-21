package scm

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/olblak/updateCli/pkg/git"
	"github.com/olblak/updateCli/pkg/github"
)

// Scm is an interface that offers common functions for a source control manager like git or github
type Scm interface {
	Add(file string)
	Clone() string
	GetDirectory() (directory string)
	Init(source string) error
	Push()
	Commit(file, message string)
	Clean()
}

// Unmarshal parses a scm struct like git or github and returns a scm interface
func Unmarshal(scm map[string]interface{}) (Scm, error) {
	var s Scm
	if len(scm) != 1 {
		return nil, fmt.Errorf("Target scm: Only one scm can be provided [git,github]")
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

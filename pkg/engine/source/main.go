package source

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/olblak/updateCli/pkg/docker"
	"github.com/olblak/updateCli/pkg/github"
	"github.com/olblak/updateCli/pkg/maven"
)

// Source defines how a value is retrieved from a specific source
type Source struct {
	Kind    string
	Output  string
	Prefix  string
	Postfix string
	Spec    interface{}
}

// Spec source is an interface to handle source spec
type Spec interface {
	Source() (string, error)
}

// Execute execute actions defined by the source configuration
func (s *Source) Execute() (string, error) {

	fmt.Printf("\n\n%s:\n", strings.ToTitle("Source"))
	fmt.Printf("%s\n\n", strings.Repeat("=", len("Source")+1))

	var output string
	var err error

	var spec Spec

	switch s.Kind {
	case "githubRelease":
		g := github.Github{}
		err := mapstructure.Decode(s.Spec, &g)

		if err != nil {
			return "", err
		}

		spec = &g

	case "maven":
		m := maven.Maven{}
		err := mapstructure.Decode(s.Spec, &m)

		if err != nil {
			return "", err
		}

		spec = &m

	case "dockerDigest":
		d := docker.Docker{}
		err := mapstructure.Decode(s.Spec, &d)

		if err != nil {
			return "", err
		}

		spec = &d

	default:
		return "", fmt.Errorf("⚠ Don't support source kind: %v", s.Kind)
	}

	output, err = spec.Source()
	if err != nil {
		return "", err
	}

	s.Output = s.Prefix + output + s.Postfix

	return s.Output, nil
}

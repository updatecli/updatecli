package source

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/olblak/updateCli/pkg/docker"
	"github.com/olblak/updateCli/pkg/github"
	"github.com/olblak/updateCli/pkg/helm/chart"
	"github.com/olblak/updateCli/pkg/maven"
)

// Source defines how a value is retrieved from a specific source
type Source struct {
	Name      string
	Kind      string
	Changelog string
	Output    string
	Prefix    string
	Postfix   string
	Replaces  Replacers
	Spec      interface{}
	Result    string `yaml:"-"` // Ignore this key field when unmarshalling yaml file
}

// Spec source is an interface to handle source spec
type Spec interface {
	Source() (string, error)
}

// Changelog is an interface to retrieve changelog description
type Changelog interface {
	Changelog(release string) (string, error)
}

// Execute execute actions defined by the source configuration
func (s *Source) Execute() error {

	fmt.Printf("\n\n%s:\n", strings.ToTitle("Source"))
	fmt.Printf("%s\n\n", strings.Repeat("=", len("Source")+1))

	var output string
	var err error

	var spec Spec
	var changelog Changelog

	switch s.Kind {
	case "githubRelease":
		g := github.Github{}
		err := mapstructure.Decode(s.Spec, &g)

		if err != nil {
			return err
		}

		spec = &g
		changelog = &g

	case "helmChart":
		c := chart.Chart{}
		err := mapstructure.Decode(s.Spec, &c)

		if err != nil {
			return err
		}

		spec = &c

	case "maven":
		m := maven.Maven{}
		err := mapstructure.Decode(s.Spec, &m)

		if err != nil {
			return err
		}

		spec = &m

	case "dockerDigest":
		d := docker.Docker{}
		err := mapstructure.Decode(s.Spec, &d)

		if err != nil {
			return err
		}

		spec = &d

	default:
		return fmt.Errorf("⚠ Don't support source kind: %v", s.Kind)
	}

	output, err = spec.Source()
	if err != nil {
		return err
	}

	if changelog != nil && s.Changelog == "" {
		s.Changelog, err = changelog.Changelog(output)
		if err != nil {
			return err
		}
	}

	if len(s.Replaces) > 0 {
		args := s.Replaces.Unmarshal()

		r := strings.NewReplacer(args...)
		s.Output = (r.Replace(output))
	} else {
		s.Output = output
	}

	fmt.Println(s.Changelog)

	return nil
}

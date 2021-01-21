package source

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/olblak/updateCli/pkg/core/scm"
	"github.com/olblak/updateCli/pkg/plugins/docker"
	"github.com/olblak/updateCli/pkg/plugins/github"
	"github.com/olblak/updateCli/pkg/plugins/helm/chart"
	"github.com/olblak/updateCli/pkg/plugins/maven"
	"github.com/olblak/updateCli/pkg/plugins/yaml"
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
	Scm       map[string]interface{}
	Result    string `yaml:"-"` // Ignore this key field when unmarshalling yaml file
}

// Spec source is an interface to handle source spec
type Spec interface {
	Source(workingDir string) (string, error)
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

	spec, changelog, err := s.Unmarshal()

	if err != nil {
		return err
	}

	workingDir := ""

	if len(s.Scm) > 0 {

		SCM, err := scm.Unmarshal(s.Scm)

		if err != nil {
			return err
		}

		err = SCM.Init("", workingDir)

		if err != nil {
			return err
		}

		err = SCM.Checkout()

		if err != nil {
			return err
		}

		workingDir = SCM.GetDirectory()

	} else if len(s.Scm) == 0 {

		pwd, err := os.Getwd()
		if err != nil {
			return err
		}

		workingDir = filepath.Dir(pwd)
	}

	output, err = spec.Source(workingDir)

	if err != nil {
		return err
	}

	if changelog != nil && s.Changelog == "" {
		s.Changelog, err = changelog.Changelog(output)
		if err != nil {
			return err
		}
	} else if changelog == nil && s.Changelog == "" {
		s.Changelog = "We couldn't identify a way to automatically retrieve changelog information"
	} else {
		return fmt.Errorf("Something weird happened while setting changelog")
	}

	if len(s.Replaces) > 0 {
		args := s.Replaces.Unmarshal()

		r := strings.NewReplacer(args...)
		s.Output = (r.Replace(output))
	} else {
		s.Output = output
	}

	if len(s.Changelog) > 0 {
		fmt.Printf("\n\n%s:\n", strings.ToTitle("Changelog"))
		fmt.Printf("%s\n", strings.Repeat("=", len("Changelog")+1))

		fmt.Printf("%s\n", s.Changelog)

	}

	return nil
}

// Unmarshal decode a source spec and returned its typed content
func (s *Source) Unmarshal() (spec Spec, changelog Changelog, err error) {
	switch s.Kind {
	case "githubRelease":
		g := github.Github{}
		err := mapstructure.Decode(s.Spec, &g)

		if err != nil {
			return nil, nil, err
		}

		spec = &g
		changelog = &g

	case "helmChart":
		c := chart.Chart{}
		err := mapstructure.Decode(s.Spec, &c)

		if err != nil {
			return nil, nil, err
		}

		spec = &c
		changelog = &c

	case "maven":
		m := maven.Maven{}
		err := mapstructure.Decode(s.Spec, &m)

		if err != nil {
			return nil, nil, err
		}

		spec = &m

	case "dockerDigest":
		d := docker.Docker{}
		err := mapstructure.Decode(s.Spec, &d)

		if err != nil {
			return nil, nil, err
		}

		spec = &d

	case "yaml":
		y := yaml.Yaml{}
		err := mapstructure.Decode(s.Spec, &y)

		if err != nil {
			return nil, nil, err
		}

		spec = &y

	default:
		return nil, nil, fmt.Errorf("⚠ Don't support source kind: %v", s.Kind)
	}
	return spec, changelog, nil

}

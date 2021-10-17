package source

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/core/scm"
	"github.com/updatecli/updatecli/pkg/core/transformer"
	"github.com/updatecli/updatecli/pkg/plugins/aws/ami"
	"github.com/updatecli/updatecli/pkg/plugins/docker"
	"github.com/updatecli/updatecli/pkg/plugins/file"
	gitTag "github.com/updatecli/updatecli/pkg/plugins/git/tag"
	"github.com/updatecli/updatecli/pkg/plugins/github"
	"github.com/updatecli/updatecli/pkg/plugins/helm/chart"
	"github.com/updatecli/updatecli/pkg/plugins/jenkins"
	"github.com/updatecli/updatecli/pkg/plugins/maven"
	"github.com/updatecli/updatecli/pkg/plugins/shell"
	yml "github.com/updatecli/updatecli/pkg/plugins/yaml"
)

// Source defines how a value is retrieved from a specific source
type Source struct {
	Changelog string // Changelog hold the changelog description
	Result    string // Result store the source result after a source run. This variable can't be set by an updatecli configuration
	Output    string // Output contains the value retrieved from a source
	Spec      Spec   // Spec defines a source specifications
}

// Spec struct defines a source specifications
type Spec struct {
	DependsOn    []string                 `yaml:"depends_on"` // DependsOn specify dag dependencies between sources
	Name         string                   // Name contains a source name
	Kind         string                   // Kind defines a source kind
	Prefix       string                   // Deprecated in favor of Transformers on 2021/01/3
	Postfix      string                   // Deprecated in favor of Transformers on 2021/01/3
	Transformers transformer.Transformers // Transformers defines the list of transformers to apply to a source Output
	Replaces     Replacers                // Deprecated in favor of Transformers on 2021/01/3
	Spec         interface{}
	Scm          map[string]interface{}
}

// Sourcer source is an interface to handle source spec
type Sourcer interface {
	Source(workingDir string) (string, error)
}

// Execute execute actions defined by the source configuration
func (s *Source) Execute() (output string, changelogContent string, err error) {

	logrus.Infof("\n\n%s:\n", strings.ToTitle("Source"))
	logrus.Infof("%s\n\n", strings.Repeat("=", len("Source")+1))

	spec, changelog, err := s.Unmarshal()

	if err != nil {
		return output, changelogContent, err
	}

	workingDir := ""

	if len(s.Spec.Scm) > 0 {

		SCM, _, err := scm.Unmarshal(s.Spec.Scm)

		if err != nil {
			return output, changelogContent, err
		}

		err = SCM.Init("", workingDir)

		if err != nil {
			return output, changelogContent, err
		}

		err = SCM.Checkout()

		if err != nil {
			return output, changelogContent, err
		}

		workingDir = SCM.GetDirectory()

	} else if len(s.Spec.Scm) == 0 {

		pwd, err := os.Getwd()
		if err != nil {
			return output, changelogContent, err
		}

		workingDir = pwd
	}

	output, err = spec.Source(workingDir)

	// Retrieve changelog using default source output before
	// modifying its value with the transformer
	if changelog != nil {
		changelogContent, err = changelog.Changelog(output)
		if err != nil {
			return output, changelogContent, err
		}
	} else if changelog == nil {
		changelogContent = "We couldn't identify a way to automatically retrieve changelog information"
	} else {
		err = fmt.Errorf("Something weird happened while setting changelog")
		return output, changelogContent, err
	}

	if len(s.Spec.Transformers) > 0 {
		output, err = s.Spec.Transformers.Apply(output)
		if err != nil {
			return output, changelogContent, err
		}
	}

	// Announce deprecation on 2021/01/31
	if len(s.Spec.Prefix) > 0 {
		logrus.Warnf("Key 'prefix' deprecated in favor of 'transformers', it will be delete in a future release\n")
	}

	// Announce deprecation on 2021/01/31
	if len(s.Spec.Postfix) > 0 {
		logrus.Warnf("Key 'postfix' deprecated in favor of 'transformers', it will be delete in a future release\n")
	}

	if err != nil {
		return output, changelogContent, err
	}

	// Deprecated in favor of Transformers on 2021/01/3
	if len(s.Spec.Replaces) > 0 {
		args := s.Spec.Replaces.Unmarshal()

		r := strings.NewReplacer(args...)
		output = (r.Replace(output))
	}

	if len(changelogContent) > 0 {
		logrus.Infof("\n\n%s:\n", strings.ToTitle("Changelog"))
		logrus.Infof("%s\n", strings.Repeat("=", len("Changelog")+1))

		logrus.Infof("%s\n", changelogContent)

	}

	return output, changelogContent, err
}

// Unmarshal decode a source spec and returned its typed content
func (s *Source) Unmarshal() (sourcer Sourcer, changelog Changelog, err error) {
	switch s.Spec.Kind {
	case "aws/ami":
		a := ami.AMI{}

		err := mapstructure.Decode(s.Spec.Spec, &a.Spec)

		if err != nil {
			return nil, nil, err
		}

		sourcer = &a

	case "githubRelease":
		githubSpec := github.Spec{}

		err := mapstructure.Decode(s.Spec.Spec, &githubSpec)

		if err != nil {
			return nil, nil, err
		}

		g, err := github.New(githubSpec)

		if err != nil {
			return nil, nil, err
		}

		sourcer = &g
		changelog = &g

	case "file":
		f := file.File{}

		err := mapstructure.Decode(s.Spec.Spec, &f)

		if err != nil {
			return nil, nil, err
		}

		sourcer = &f

	case "helmChart":
		c := chart.Chart{}
		err := mapstructure.Decode(s.Spec.Spec, &c)

		if err != nil {
			return nil, nil, err
		}

		sourcer = &c
		changelog = &c

	case "jenkins":
		j := jenkins.Jenkins{}

		err := mapstructure.Decode(s.Spec.Spec, &j)

		if err != nil {
			return nil, nil, err
		}

		sourcer = &j
		changelog = &j

	case "maven":
		m := maven.Maven{}
		err := mapstructure.Decode(s.Spec.Spec, &m)

		if err != nil {
			return nil, nil, err
		}

		sourcer = &m

	case "gitTag":
		g := gitTag.Tag{}
		err := mapstructure.Decode(s.Spec.Spec, &g)

		if err != nil {
			return nil, nil, err
		}

		sourcer = &g

	case "dockerDigest":
		d := docker.Docker{}
		err := mapstructure.Decode(s.Spec.Spec, &d)

		if err != nil {
			return nil, nil, err
		}

		sourcer = &d

	case "yaml":
		y := yml.Yaml{}
		err := mapstructure.Decode(s.Spec.Spec, &y)

		if err != nil {
			return nil, nil, err
		}

		sourcer = &y

	case "shell":
		shellResourceSpec := shell.ShellSpec{}

		err := mapstructure.Decode(s.Spec.Spec, &shellResourceSpec)
		if err != nil {
			return nil, nil, err
		}

		sourcer, err = shell.New(shellResourceSpec)
		if err != nil {
			return nil, nil, err
		}

	default:
		return nil, nil, fmt.Errorf("⚠ Don't support source kind: %v", s.Spec.Kind)
	}
	return sourcer, changelog, nil

}

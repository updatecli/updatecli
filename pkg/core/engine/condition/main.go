package condition

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/olblak/updateCli/pkg/core/scm"
	"github.com/olblak/updateCli/pkg/core/transformer"
	"github.com/olblak/updateCli/pkg/plugins/docker"
	"github.com/olblak/updateCli/pkg/plugins/docker/dockerfile"
	"github.com/olblak/updateCli/pkg/plugins/file"
	gitTag "github.com/olblak/updateCli/pkg/plugins/git/tag"
	"github.com/olblak/updateCli/pkg/plugins/helm/chart"
	"github.com/olblak/updateCli/pkg/plugins/jenkins"
	"github.com/olblak/updateCli/pkg/plugins/maven"
	"github.com/olblak/updateCli/pkg/plugins/yaml"
	"github.com/sirupsen/logrus"
)

// Condition defines which condition needs to be met
// in order to update targets based on the source output
type Condition struct {
	DependsOn    []string `yaml:"depends_on"`
	Name         string
	Kind         string
	Prefix       string // Deprecated in favor of Transformers on 2021/01/3
	Postfix      string // Deprecated in favor of Transformers on 2021/01/3
	Transformers transformer.Transformers
	Spec         interface{}
	Scm          map[string]interface{}
	SourceID     string `yaml:"sourceID"`
}

// Spec is an interface that test if condition is met
type Spec interface {
	Condition(version string) (bool, error)
	ConditionFromSCM(version string, scm scm.Scm) (bool, error)
}

// Run tests if a specific condition is true
func (c *Condition) Run(source string) (ok bool, err error) {
	ok = true

	spec, err := Unmarshal(c)
	if err != nil {
		return false, err
	}

	if len(c.Transformers) > 0 {
		source, err = c.Transformers.Apply(source)
		if err != nil {
			return false, err
		}
	}

	// Announce deprecation on 2021/01/31
	if len(c.Prefix) > 0 {
		logrus.Warnf("Key 'prefix' deprecated in favor of 'transformers', it will be delete in a future release")
	}

	// Announce deprecation on 2021/01/31
	if len(c.Postfix) > 0 {
		logrus.Warnf("Key 'postfix' deprecated in favor of 'transformers', it will be delete in a future release")
	}

	// If scm is defined then clone the repository
	if len(c.Scm) > 0 {
		s, _, err := scm.Unmarshal(c.Scm)
		if err != nil {
			return false, err
		}

		err = s.Init(c.Prefix+source+c.Postfix, c.Name)
		if err != nil {
			return false, err
		}

		err = s.Checkout()
		if err != nil {
			return false, err
		}

		ok, err = spec.ConditionFromSCM(c.Prefix+source+c.Postfix, s)
		if err != nil {
			return false, err
		}

	} else if len(c.Scm) == 0 {
		ok, err = spec.Condition(c.Prefix + source + c.Postfix)
		if err != nil {
			return false, err
		}
	} else {
		return false, fmt.Errorf("Something went wrong while looking at the scm configuration: %v", c.Scm)
	}

	return ok, nil

}

// Unmarshal decodes a condition struct
func Unmarshal(condition *Condition) (spec Spec, err error) {

	switch condition.Kind {

	case "dockerImage":
		d := docker.Docker{}

		err := mapstructure.Decode(condition.Spec, &d)
		if err != nil {
			return nil, err
		}

		spec = &d

	case "dockerfile":
		d := dockerfile.Dockerfile{}

		err := mapstructure.Decode(condition.Spec, &d)
		if err != nil {
			return nil, err
		}

		spec = &d

	case "file":
		f := file.File{}

		err := mapstructure.Decode(condition.Spec, &f)
		if err != nil {
			return nil, err
		}

		spec = &f

	case "jenkins":
		j := jenkins.Jenkins{}

		err := mapstructure.Decode(condition.Spec, &j)
		if err != nil {
			return nil, err
		}

		spec = &j

	case "maven":
		m := maven.Maven{}

		err := mapstructure.Decode(condition.Spec, &m)
		if err != nil {
			return nil, err
		}

		spec = &m

	case "gitTag":
		g := gitTag.Tag{}
		err := mapstructure.Decode(condition.Spec, &g)

		if err != nil {
			return nil, err
		}

		spec = &g

	case "helmChart":
		ch := chart.Chart{}

		err := mapstructure.Decode(condition.Spec, &ch)
		if err != nil {
			return nil, err
		}

		spec = &ch

	case "yaml":
		y := yaml.Yaml{}

		err := mapstructure.Decode(condition.Spec, &y)
		if err != nil {
			return nil, err
		}

		spec = &y

	default:
		return nil, fmt.Errorf("Don't support condition: %v", condition.Kind)
	}
	return spec, nil
}

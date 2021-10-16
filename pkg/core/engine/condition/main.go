package condition

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/scm"
	"github.com/updatecli/updatecli/pkg/core/transformer"
	"github.com/updatecli/updatecli/pkg/plugins/aws/ami"
	"github.com/updatecli/updatecli/pkg/plugins/docker"
	"github.com/updatecli/updatecli/pkg/plugins/docker/dockerfile"
	"github.com/updatecli/updatecli/pkg/plugins/file"
	gitTag "github.com/updatecli/updatecli/pkg/plugins/git/tag"
	"github.com/updatecli/updatecli/pkg/plugins/helm/chart"
	"github.com/updatecli/updatecli/pkg/plugins/jenkins"
	"github.com/updatecli/updatecli/pkg/plugins/maven"
	"github.com/updatecli/updatecli/pkg/plugins/shell"
	yml "github.com/updatecli/updatecli/pkg/plugins/yaml"
	"gopkg.in/yaml.v3"
)

// Condition defines which condition needs to be met
// in order to update targets based on the source output
type Condition struct {
	Result string // Result store the condition result after a condition run. This variable can't be set by an updatecli configuration
	Spec   Spec   // Spec defines condition input parameters
}

// Spec defines conditions input parameters
type Spec struct {
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

// Conditioner is an interface that test if condition is met
type Conditioner interface {
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

	if len(c.Spec.Transformers) > 0 {
		source, err = c.Spec.Transformers.Apply(source)
		if err != nil {
			return false, err
		}
	}

	// Announce deprecation on 2021/01/31
	if len(c.Spec.Prefix) > 0 {
		logrus.Warnf("Key 'prefix' deprecated in favor of 'transformers', it will be delete in a future release")
	}

	// Announce deprecation on 2021/01/31
	if len(c.Spec.Postfix) > 0 {
		logrus.Warnf("Key 'postfix' deprecated in favor of 'transformers', it will be delete in a future release")
	}

	// If scm is defined then clone the repository
	if len(c.Spec.Scm) > 0 {
		s, _, err := scm.Unmarshal(c.Spec.Scm)
		if err != nil {
			return false, err
		}

		err = s.Init(c.Spec.Prefix+source+c.Spec.Postfix, c.Spec.Name)
		if err != nil {
			return false, err
		}

		err = s.Checkout()
		if err != nil {
			return false, err
		}

		ok, err = spec.ConditionFromSCM(c.Spec.Prefix+source+c.Spec.Postfix, s)
		if err != nil {
			return false, err
		}

	} else if len(c.Spec.Scm) == 0 {
		ok, err = spec.Condition(c.Spec.Prefix + source + c.Spec.Postfix)
		if err != nil {
			return false, err
		}
	} else {
		return false, fmt.Errorf("Something went wrong while looking at the scm configuration: %v", c.Spec.Scm)
	}

	return ok, nil

}

// Unmarshal decodes a condition struct
func Unmarshal(condition *Condition) (conditioner Conditioner, err error) {

	switch condition.Spec.Kind {

	case "aws/ami":
		a := ami.AMI{}

		err := mapstructure.Decode(condition.Spec.Spec, &a.Spec)

		if err != nil {
			return nil, err
		}

		conditioner = &a

	case "dockerImage":
		d := docker.Docker{}

		err := mapstructure.Decode(condition.Spec.Spec, &d)
		if err != nil {
			return nil, err
		}

		conditioner = &d

	case "dockerfile":
		d := dockerfile.Dockerfile{}

		err := mapstructure.Decode(condition.Spec.Spec, &d)
		if err != nil {
			return nil, err
		}

		conditioner = &d

	case "file":
		f := file.File{}

		err := mapstructure.Decode(condition.Spec.Spec, &f)
		if err != nil {
			return nil, err
		}

		conditioner = &f

	case "jenkins":
		j := jenkins.Jenkins{}

		err := mapstructure.Decode(condition.Spec.Spec, &j)
		if err != nil {
			return nil, err
		}

		conditioner = &j

	case "maven":
		m := maven.Maven{}

		err := mapstructure.Decode(condition.Spec.Spec, &m)
		if err != nil {
			return nil, err
		}

		conditioner = &m

	case "gitTag":
		g := gitTag.Tag{}
		err := mapstructure.Decode(condition.Spec.Spec, &g)

		if err != nil {
			return nil, err
		}

		conditioner = &g

	case "helmChart":
		ch := chart.Chart{}

		err := mapstructure.Decode(condition.Spec.Spec, &ch)
		if err != nil {
			return nil, err
		}

		conditioner = &ch

	case "yaml":
		y := yml.Yaml{}

		err := mapstructure.Decode(condition.Spec.Spec, &y)
		if err != nil {
			return nil, err
		}

		conditioner = &y

	case "shell":
		var shellResourceSpec shell.ShellSpec

		if err := mapstructure.Decode(condition.Spec.Spec, &shellResourceSpec); err != nil {
			return nil, err
		}

		conditioner, err = shell.New(shellResourceSpec)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("Don't support condition: %v", condition.Spec.Kind)
	}
	return conditioner, nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (c *Condition) UnmarshalYAML(value *yaml.Node) error {

	var spec Spec

	if err := value.Decode(&spec); err != nil {
		logrus.Errorln(err)
		return err
	}

	c.Spec = spec

	return nil
}

// MarshalYAML implements the yaml.Unmarshaler interface.
// https://github.com/go-yaml/yaml/issues/714
func (c Condition) MarshalYAML() (interface{}, error) {
	node := yaml.Node{}
	err := node.Encode(c.Spec)
	if err != nil {
		return nil, err
	}
	return node, nil
}

package condition

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
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
	"github.com/updatecli/updatecli/pkg/plugins/yaml"
)

// Condition defines which condition needs to be met
// in order to update targets based on the source output
type Condition struct {
	Result string // Result store the condition result after a condition run. This variable can't be set by an updatecli configuration
	Config Config // Config defines condition input parameters
	Scm    *scm.ScmHandler
}

// Config defines conditions input parameters
type Config struct {
	DependsOn          []string `yaml:"depends_on"`
	Name               string
	Kind               string
	Prefix             string // Deprecated in favor of Transformers on 2021/01/3
	Postfix            string // Deprecated in favor of Transformers on 2021/01/3
	Transformers       transformer.Transformers
	Spec               interface{}
	Scm                map[string]interface{} // Deprecated field on version [x.y.z]
	SCMID              string                 `yaml:"scmID"` // SCMID references a uniq scm configuration
	SourceID           string                 `yaml:"sourceID"`
	DisableSourceInput bool
}

// Conditioner is an interface that test if condition is met
type Conditioner interface {
	Condition(version string) (bool, error)
	ConditionFromSCM(version string, scm scm.ScmHandler) (bool, error)
}

// Run tests if a specific condition is true
func (c *Condition) Run(source string) (err error) {
	ok := false

	spec, err := Unmarshal(c)
	if err != nil {
		c.Result = result.FAILURE
		return err
	}

	if len(c.Config.Transformers) > 0 {
		source, err = c.Config.Transformers.Apply(source)
		if err != nil {
			c.Result = result.FAILURE
			return err
		}
	}

	// Announce deprecation on 2021/01/31
	if len(c.Config.Prefix) > 0 {
		logrus.Warnf("Key 'prefix' deprecated in favor of 'transformers', it will be delete in a future release")
	}

	// Announce deprecation on 2021/01/31
	if len(c.Config.Postfix) > 0 {
		logrus.Warnf("Key 'postfix' deprecated in favor of 'transformers', it will be delete in a future release")
	}

	// If scm is defined then clone the repository
	if c.Scm != nil {
		s := *c.Scm
		if err != nil {
			c.Result = result.FAILURE
			return err
		}

		err = s.Init(c.Config.Prefix+source+c.Config.Postfix, c.Config.Name)
		if err != nil {
			c.Result = result.FAILURE
			return err
		}

		err = s.Checkout()
		if err != nil {
			c.Result = result.FAILURE
			return err
		}

		ok, err = spec.ConditionFromSCM(c.Config.Prefix+source+c.Config.Postfix, s)
		if err != nil {
			c.Result = result.FAILURE
			return err
		}

	} else if len(c.Config.Scm) == 0 {
		ok, err = spec.Condition(c.Config.Prefix + source + c.Config.Postfix)
		if err != nil {
			c.Result = result.FAILURE
			return err
		}
	} else {
		return fmt.Errorf("something went wrong while looking at the scm configuration: %v", c.Config.Scm)
	}

	if ok {
		c.Result = result.SUCCESS
	} else {
		c.Result = result.FAILURE
	}

	return nil

}

// Unmarshal decodes a condition struct
func Unmarshal(condition *Condition) (conditioner Conditioner, err error) {

	switch condition.Config.Kind {

	case "aws/ami":
		a := ami.AMI{}

		err := mapstructure.Decode(condition.Config.Spec, &a.Spec)

		if err != nil {
			return nil, err
		}

		conditioner = &a

	case "dockerImage":
		d := docker.Docker{}

		err := mapstructure.Decode(condition.Config.Spec, &d)
		if err != nil {
			return nil, err
		}

		conditioner = &d

	case "dockerfile":
		d := dockerfile.Dockerfile{}

		err := mapstructure.Decode(condition.Config.Spec, &d)
		if err != nil {
			return nil, err
		}

		conditioner = &d

	case "file":
		var conditionSpec file.FileSpec

		if err := mapstructure.Decode(condition.Config.Spec, &conditionSpec); err != nil {
			return nil, err
		}

		conditioner, err = file.New(conditionSpec)
		if err != nil {
			return nil, err
		}

	case "jenkins":
		j := jenkins.Jenkins{}

		err := mapstructure.Decode(condition.Config.Spec, &j)
		if err != nil {
			return nil, err
		}

		conditioner = &j

	case "maven":
		m := maven.Maven{}

		err := mapstructure.Decode(condition.Config.Spec, &m)
		if err != nil {
			return nil, err
		}

		conditioner = &m

	case "gitTag":
		g := gitTag.Tag{}
		err := mapstructure.Decode(condition.Config.Spec, &g)

		if err != nil {
			return nil, err
		}

		conditioner = &g

	case "helmChart":
		ch := chart.Chart{}

		err := mapstructure.Decode(condition.Config.Spec, &ch)
		if err != nil {
			return nil, err
		}

		conditioner = &ch

	case "yaml":
		var conditionSpec yaml.YamlSpec

		if err := mapstructure.Decode(condition.Config.Spec, &conditionSpec); err != nil {
			return nil, err
		}

		conditioner, err = yaml.New(conditionSpec)
		if err != nil {
			return nil, err
		}

	case "shell":
		var conditionSpec shell.ShellSpec

		if err := mapstructure.Decode(condition.Config.Spec, &conditionSpec); err != nil {
			return nil, err
		}

		conditioner, err = shell.New(conditionSpec)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("Don't support condition: %v", condition.Config.Kind)
	}
	return conditioner, nil
}

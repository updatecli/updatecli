package condition

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/transformer"
	"github.com/updatecli/updatecli/pkg/plugins/awsami"
	"github.com/updatecli/updatecli/pkg/plugins/dockerfile"
	"github.com/updatecli/updatecli/pkg/plugins/dockerimage"
	"github.com/updatecli/updatecli/pkg/plugins/file"
	"github.com/updatecli/updatecli/pkg/plugins/gittag"
	"github.com/updatecli/updatecli/pkg/plugins/helm"
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
	Scm                map[string]interface{} // Deprecated field on version [1.17.0]
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
		logrus.Errorf("%s", err)
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
		var conditionSpec awsami.Spec

		if err := mapstructure.Decode(condition.Config.Spec, &conditionSpec); err != nil {
			return nil, err
		}

		conditioner, err = awsami.New(conditionSpec)
		if err != nil {
			return nil, err
		}

	case "dockerImage":
		var conditionSpec dockerimage.Spec

		if err := mapstructure.Decode(condition.Config.Spec, &conditionSpec); err != nil {
			return nil, err
		}

		conditioner, err = dockerimage.New(conditionSpec)
		if err != nil {
			return nil, err
		}

	case "dockerfile":
		var conditionSpec dockerfile.Spec

		if err := mapstructure.Decode(condition.Config.Spec, &conditionSpec); err != nil {
			return nil, err
		}

		conditioner, err = dockerfile.New(conditionSpec)
		if err != nil {
			return nil, err
		}

	case "file":
		var conditionSpec file.Spec

		if err := mapstructure.Decode(condition.Config.Spec, &conditionSpec); err != nil {
			return nil, err
		}

		conditioner, err = file.New(conditionSpec)
		if err != nil {
			return nil, err
		}

	case "jenkins":
		var conditionSpec jenkins.Spec

		if err := mapstructure.Decode(condition.Config.Spec, &conditionSpec); err != nil {
			return nil, err
		}

		conditioner, err = jenkins.New(conditionSpec)
		if err != nil {
			return nil, err
		}

	case "maven":
		var conditionSpec maven.Spec

		if err := mapstructure.Decode(condition.Config.Spec, &conditionSpec); err != nil {
			return nil, err
		}

		conditioner, err = maven.New(conditionSpec)
		if err != nil {
			return nil, err
		}

	case "gitTag":
		var conditionSpec gittag.Spec

		if err := mapstructure.Decode(condition.Config.Spec, &conditionSpec); err != nil {
			return nil, err
		}

		conditioner, err = gittag.New(conditionSpec)
		if err != nil {
			return nil, err
		}

	case "helmChart":
		var conditionSpec helm.Spec

		if err := mapstructure.Decode(condition.Config.Spec, &conditionSpec); err != nil {
			return nil, err
		}

		conditioner, err = helm.New(conditionSpec)
		if err != nil {
			return nil, err
		}

	case "yaml":
		var conditionSpec yaml.Spec

		if err := mapstructure.Decode(condition.Config.Spec, &conditionSpec); err != nil {
			return nil, err
		}

		conditioner, err = yaml.New(conditionSpec)
		if err != nil {
			return nil, err
		}

	case "shell":
		var conditionSpec shell.Spec

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

package condition

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
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
	resource.ResourceConfig `yaml:",inline"`
	SourceID                string `yaml:"sourceID"`
	DisableSourceInput      bool
}

// Run tests if a specific condition is true
func (c *Condition) Run(source string) (err error) {
	ok := false

	condition, err := resource.New(c.Config.ResourceConfig)
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

		ok, err = condition.ConditionFromSCM(c.Config.Prefix+source+c.Config.Postfix, s)
		if err != nil {
			c.Result = result.FAILURE
			return err
		}

	} else if len(c.Config.Scm) == 0 {
		ok, err = condition.Condition(c.Config.Prefix + source + c.Config.Postfix)
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

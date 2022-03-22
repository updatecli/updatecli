package condition

import (
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/schema"
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
	// SourceID defines which source is used to retrieve the default value
	SourceID string `yaml:"sourceID"`
	// DisableSourceInput allows to not retrieve default source value.
	DisableSourceInput bool
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

	// If scm is defined then clone the repository
	if c.Scm != nil {
		s := *c.Scm
		if err != nil {
			c.Result = result.FAILURE
			return err
		}

		err = s.Init(c.Config.Name)
		if err != nil {
			c.Result = result.FAILURE
			return err
		}

		err = s.Checkout()
		if err != nil {
			c.Result = result.FAILURE
			return err
		}

		ok, err = condition.ConditionFromSCM(source, s)
		if err != nil {
			c.Result = result.FAILURE
			return err
		}

	} else if len(c.Config.Scm) == 0 {
		ok, err = condition.Condition(source)
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

// JSONSchema implements the json schema interface to generate condition json schema.
func (c Config) JSONSchema() *jsonschema.Schema {

	anyOfSpec := resource.GetResourceMapping()

	return schema.GenerateJsonSchema(Config{}, anyOfSpec)
}

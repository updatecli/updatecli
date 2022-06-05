package condition

import (
	"fmt"

	jschema "github.com/invopop/jsonschema"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/jsonschema"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition defines which condition needs to be met
// in order to update targets based on the source output
type Condition struct {
	// Result stores the condition result after a condition run.
	Result string
	// Config defines condition input parameters
	Config Config
	Scm    *scm.ScmHandler
}

// Config defines conditions input parameters
type Config struct {
	resource.ResourceConfig `yaml:",inline"`
	// ! Deprecated in favor of sourceID
	// sourceid specifies which "source", based on its ID, is used to retrieve the default value.
	DeprecatedSourceID string `yaml:"sourceID"`
	// sourceid specifies which "source", based on its ID, is used to retrieve the default value.
	SourceID string
	// disablesourceinput disable the mechanism to retrieve a default value from a source.
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
		var i interface{} = c.Config.Scm
		scm := i.(scm.ScmHandler)
		return fmt.Errorf("something went wrong while looking at the scm configuration: %s", scm.ToString())
	}

	if ok {
		c.Result = result.SUCCESS
	} else {
		c.Result = result.FAILURE
	}

	return nil

}

// JSONSchema implements the json schema interface to generate the "condition" jsonschema.
func (c Config) JSONSchema() *jschema.Schema {

	type configAlias Config
	anyOfSpec := resource.GetResourceMapping()

	return jsonschema.GenerateJsonSchema(configAlias{}, anyOfSpec)
}

func (c *Config) Validate() error {
	// Handle scmID deprecation
	if len(c.DeprecatedSCMID) > 0 {
		switch len(c.SCMID) {
		case 0:
			logrus.Warningf("%q is deprecated in favor of %q.", "scmID", "scmid")
			c.SCMID = c.DeprecatedSCMID
			c.DeprecatedSCMID = ""
		default:
			logrus.Warningf("%q and %q are mutually exclusif, ignoring %q",
				"scmID", "scmid", "scmID")
		}
	}

	// Handle sourceID deprecation
	if len(c.DeprecatedSourceID) > 0 {
		switch len(c.SourceID) {
		case 0:
			logrus.Warningf("%q is deprecated in favor of %q.", "sourceID", "sourceid")
			c.SourceID = c.DeprecatedSourceID
			c.DeprecatedSourceID = ""
		default:
			logrus.Warningf("%q and %q are mutually exclusif, ignoring %q",
				"sourceID", "sourceid", "sourceID")
		}
	}

	return nil
}

package condition

import (
	"errors"
	"strings"

	jschema "github.com/invopop/jsonschema"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/jsonschema"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

var (
	// ErrWrongConfig is returned when a condition spec has missing attributes which are mandatory
	ErrWrongConfig = errors.New("wrong condition configuration")
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
	resource.ResourceConfig `yaml:",inline,omitempty"`
	// ! Deprecated in favor of sourceID
	DeprecatedSourceID string `yaml:"sourceID,omitempty" jsonschema:"-"`
	// sourceid specifies which "source", based on its ID, is used to retrieve the default value.
	SourceID string `yaml:",omitempty"`
	// disablesourceinput disable the mechanism to retrieve a default value from a source.
	DisableSourceInput bool `yaml:",omitempty"`
}

// Run tests if a specific condition is true
func (c *Condition) Run(source string) (err error) {
	c.Result = result.FAILURE
	ok := false

	condition, err := resource.New(c.Config.ResourceConfig)
	if err != nil {
		return err
	}

	if len(c.Config.Transformers) > 0 {
		source, err = c.Config.Transformers.Apply(source)
		if err != nil {
			return err
		}
	}

	switch c.Scm == nil {
	case true:
		ok, err = condition.Condition(source)
		if err != nil {
			return err
		}
	case false:
		// If scm is defined then clone the repository
		s := *c.Scm
		if err != nil {
			return err
		}

		err = s.Checkout()
		if err != nil {
			return err
		}

		ok, err = condition.ConditionFromSCM(source, s)
		if err != nil {
			return err
		}
	}

	if ok {
		c.Result = result.SUCCESS
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
	gotError := false
	missingParameters := []string{}

	// Validate that kind is set
	if len(c.Kind) == 0 {
		missingParameters = append(missingParameters, "kind")
	}

	// Ensure kind is lowercase
	if c.Kind != strings.ToLower(c.Kind) {
		logrus.Warningf("kind value %q must be lowercase", c.Kind)
		c.Kind = strings.ToLower(c.Kind)
	}

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

	// Handle depends_on deprecation
	if len(c.DeprecatedDependsOn) > 0 {
		switch len(c.DependsOn) == 0 {
		case true:
			logrus.Warningln("\"depends_on\" is deprecated in favor of \"dependson\".")
			c.DependsOn = c.DeprecatedDependsOn
			c.DeprecatedDependsOn = []string{}
		case false:
			logrus.Warningln("\"depends_on\" is ignored in favor of \"dependson\".")
			c.DeprecatedDependsOn = []string{}
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

	err := c.Transformers.Validate()
	if err != nil {
		return err
	}

	if len(c.SourceID) > 0 && c.DisableSourceInput {
		logrus.Errorln("disablesourceinput is incompatible with sourceid, ignoring the latter")
		gotError = true
	}

	if len(missingParameters) > 0 {
		logrus.Errorf("missing value for parameter(s) [%q]", strings.Join(missingParameters, ","))
		gotError = true
	}

	if gotError {
		return ErrWrongConfig
	}

	return nil
}

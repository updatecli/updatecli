package target

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	jschema "github.com/invopop/jsonschema"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/jsonschema"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

var (
	// ErrWrongConfig is returned when a target spec has missing attributes which are mandatory
	ErrWrongConfig = errors.New("wrong target configuration")
)

// Target defines which file needs to be updated based on source output
type Target struct {
	// Result store the condition result after a target run.
	Result *result.Target
	// Config defines target input parameters
	Config Config
	// Commit defines if a target was executed in Commit mode
	Commit bool
	// Push defines if a target was executed in Push mode
	Push bool
	// Clean defines if a target was executed in Clean mode
	Clean bool
	// DryRun defines if a target was executed in DryRun mode
	DryRun bool
	// Scm stores scm information
	Scm *scm.ScmHandler
}

// Config defines target parameters
type Config struct {
	// ResourceConfig defines target input parameters
	resource.ResourceConfig `yaml:",inline"`
	// dependsonchange enables the mechanism to check if the dependant target(s) have made a change.
	// If the dependant target(s) have not made a change the target will be skipped.
	//
	// default:
	//   false
	DependsOnChange bool `yaml:",omitempty"`
	// ! Deprecated - please use all lowercase `sourceid`
	DeprecatedSourceID string `yaml:"sourceID,omitempty" jsonschema:"-"`
	// disablesourceinput disables the mechanism to retrieve a default value from a source.
	// For example, if true, source information like changelog will not be accessible for a github/pullrequest action.
	//
	// default:
	//  false
	DisableSourceInput bool `yaml:",omitempty"`
	// sourceid specifies where retrieving the default value.
	//
	// default:
	//   if only one source is defined, then sourceid is set to that sourceid.
	SourceID string `yaml:",omitempty"`
	// ! Deprecated - please use DependsOn with `condition#conditionid` keys
	//
	// conditionids specifies the list of conditions to be evaluated before running the target.
	// if at least one condition is not met, the target will be skipped.
	//
	// default:
	//   by default, all conditions are evaluated.
	DeprecatedConditionIDs []string `yaml:"conditionids,omitempty"`
	// disableconditions disables the mechanism to evaluate all conditions before running the target.
	//
	// default:
	//   false
	//
	// remark:
	//  It's possible to only monitor specific conditions by setting disableconditions to true
	//  and using DependsOn with `condition#conditionid` keys
	DisableConditions bool `yaml:"disableconditions,omitempty"`
}

// Check verifies if mandatory Targets parameters are provided and return false if not.
func (t *Target) Check() (bool, error) {
	ok := true
	required := []string{}

	if t.Config.Name == "" {
		required = append(required, "Name")
	}

	if len(required) > 0 {
		err := fmt.Errorf("%s Target parameter(s) required: [%v]", result.FAILURE, strings.Join(required, ","))
		return false, err
	}

	return ok, nil
}

// Run applies a specific target configuration
func (t *Target) Run(source string, o *Options) (err error) {
	var consoleOutput bytes.Buffer
	// By default logrus logs to stderr, so I guess we want to keep this behavior...
	logrus.SetOutput(io.MultiWriter(os.Stdout, &consoleOutput))
	/*
		The last defer will be executed first,
		so in this case we want to first save the console output
		before setting back the logrus output to stdout.
	*/
	// By default logrus logs to stdout and we want to keep this behavior...
	defer logrus.SetOutput(os.Stdout)
	defer t.Result.SetConsoleOutput(&consoleOutput)

	failTargetRun := func() {
		t.Result.Result = result.FAILURE
		t.Result.Description = "something went wrong during pipeline execution"
	}

	if len(t.Config.Transformers) > 0 {
		source, err = t.Config.Transformers.Apply(source)
		if err != nil {
			failTargetRun()
			return err
		}
	}

	target, err := resource.New(t.Config.ResourceConfig)
	if err != nil {
		failTargetRun()
		return err
	}

	// If no scm configuration provided then stop early
	if t.Scm == nil {
		err = target.Target(source, nil, o.DryRun, t.Result)
		if err != nil {
			failTargetRun()
			return err
		}

		// Could be improve to show attention description in yellow, success in green, failure in red
		logrus.Infof("%s - %s", t.Result.Result, t.Result.Description)

		return nil
	}

	_, err = t.Check()
	if err != nil {
		failTargetRun()
		return err
	}

	s := *t.Scm

	if err = s.Checkout(); err != nil {
		failTargetRun()
		return err
	}

	err = target.Target(source, s, o.DryRun, t.Result)
	if err != nil {
		failTargetRun()
		return err
	}

	// Could be improve to show attention description in yellow, success in green, failure in red
	logrus.Infof("%s - %s", t.Result.Result, t.Result.Description)

	isRemoteBranchUpToDate, err := s.IsRemoteBranchUpToDate()
	if err != nil {
		failTargetRun()
		return err
	}

	// targetCommit is used to avoid committing changes when the target has no changes.
	targetCommit := true
	if !t.Result.Changed {
		if isRemoteBranchUpToDate {
			return nil
		}

		logrus.Infof("\n\u26A0 While nothing change in the current pipeline run, according to the git history, some commits must be pushed\n")
		t.Result.Description = fmt.Sprintf("%s\n\n%s", t.Result.Description, "While nothing change in the current pipeline run, according to the git history, some commits must pushed")

		// Even though the target has no changes, it has something to commit.
		// We consider this result as "success" and not "attention" as the target has no changes.
		// If later we decide to consider the result as "attention" then we also need to consider that the action
		// will be trigger in priority. cfr https://github.com/updatecli/updatecli/issues/2039
		// So we need to create a new resource stage to handle this case.
		t.Result.Result = result.SUCCESS
		t.Result.Changed = true
		// Even though the target has left over changes, it has nothing to commit.
		targetCommit = false
	}

	if !o.DryRun {
		// o.Commit represents Global updatecli commit option
		// targetCommit represents the local target commit option
		if o.Commit && targetCommit {
			if t.Result.Description == "" {
				failTargetRun()
				return fmt.Errorf("target has no change message")
			}

			if len(t.Result.Files) == 0 {
				failTargetRun()
				return fmt.Errorf("no changed file to commit")
			}

			if err := s.Add(t.Result.Files); err != nil {
				failTargetRun()
				return err
			}

			/*
				not every target have a name as it wasn't mandatory in the past
				so we use the description as a fallback
			*/
			commitMessage := t.Config.Name
			if commitMessage == "" {
				commitMessage = t.Result.Description
			}
			if err = s.Commit(commitMessage); err != nil {
				failTargetRun()
				return err
			}
		}

		if o.Push {
			t.Result.Scm.BranchReset, err = s.Push()
			if err != nil {
				failTargetRun()
				return err
			}
		}
	}

	return nil
}

// JSONSchema implements the json schema interface to generate the "target" jsonschema.
func (Config) JSONSchema() *jschema.Schema {
	type configAlias Config

	anyOfSpec := resource.GetResourceMapping()

	return jsonschema.AppendOneOfToJsonSchema(configAlias{}, anyOfSpec)
}

// Validate checks if a target configuration is valid
func (c *Config) Validate() error {
	// Handle scmID deprecation

	gotError := false

	missingParameters := []string{}

	// Validate that kind is set
	if len(c.Kind) == 0 {
		missingParameters = append(missingParameters, "kind")
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

	// Ensure kind is lowercase
	if c.Kind != strings.ToLower(c.Kind) {
		logrus.Warningf("kind value %q must be lowercase", c.Kind)
		c.Kind = strings.ToLower(c.Kind)
	}

	if len(c.DeprecatedSCMID) > 0 {
		switch len(c.SCMID) {
		case 0:
			logrus.Warningf("%q is deprecated in favor of %q.", "scmID", "scmid")
			c.SCMID = c.DeprecatedSCMID
			c.DeprecatedSCMID = ""
		default:
			logrus.Warningf("%q and %q are mutually exclusive, ignoring %q",
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
			logrus.Warningf("%q and %q are mutually exclusive, ignoring %q",
				"sourceID", "sourceid", "sourceID")
		}
	}

	// Handle ConditionIDs deprecation
	if len(c.DeprecatedConditionIDs) > 0 {
		if len(c.DependsOn) > 0 {
			logrus.Warningf("%q and %q are mutually exclusive, ignoring %q", "conditionids", "dependson", "conditionids")
		} else {
			logrus.Warningf("%q is deprecated in favor of %q", "conditionids", "dependson")
			for _, condition := range c.DeprecatedConditionIDs {
				logrus.Warningf("%q is deprecated in favor of %q: %s", "conditionids", "dependson", condition)
				c.DependsOn = append(c.DependsOn, fmt.Sprintf("condition#%s", condition))
			}
			c.DeprecatedConditionIDs = []string{}
			c.DisableConditions = true
		}
	}

	err := c.Transformers.Validate()
	if err != nil {
		logrus.Errorln(err)
		gotError = true
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

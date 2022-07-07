package shell

import (
	"fmt"
	"os"
)

const (
	// DryRunVariableName specifies the environment variable used within shell script to detect if we are in dryrun mode
	DryRunVariableName = "DRY_RUN"
	// CurrentStageVariableName is the environment variable containing the current pipeline stage such as source, condition, target
	CurrentStageVariableName = "UPDATECLI_PIPELINE_STAGE"
)

// Environment is a struct containing information for an environment variable such as its name and its value
type Environment struct {
	// Name defines the environment variable name
	Name string `yaml:",omitempty" jsonschema:"required"`
	// Value defines the environment variable value
	Value string `yaml:",omitempty"`
}

func (e Environment) String() string {
	return fmt.Sprintf("%s=%s", e.Name, e.Value)
}

// Load updates the environment value based on Updatecli environment variables, if the value is not defined yet
func (e *Environment) Load() error {
	err := e.Validate()
	if err != nil {
		return err
	}

	// If an environment variable name is specified and specified without value
	// then it inherits the value from Updatecli process environment
	if len(e.Value) == 0 && len(e.Name) > 0 {
		value, found := os.LookupEnv(e.Name)

		if !found {
			return fmt.Errorf("environment variable %q not found, skipping", e.Name)
		}
		e.Value = value
	}
	return nil
}

// Validate ensures that required parameter are set
func (e *Environment) Validate() error {

	if len(e.Name) == 0 {
		return fmt.Errorf("parameter %q is required", "name")
	}
	return nil
}

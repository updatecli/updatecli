package shell

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

const (
	// DryRunVariableName specifies the environment variable used within shell script to detect if we are in dryrun mode
	DryRunVariableName = "DRY_RUN"
	// CurrentStageVariableName is the environment variable containing the current pipeline stage such as source, condition, target
	CurrentStageVariableName = "UPDATECLI_PIPELINE_STAGE"
)

type Environment struct {
	// Name defines the environment variable name
	Name string `yaml:",omitempty"`
	// Value defines the environment variable value
	Value string `yaml:",omitempty"`
}

func (e Environment) String() string {
	return fmt.Sprintf("%s=%s", e.Name, e.Value)
}

// Update updates the environment value based on Updatecli environment variables, if the value is not defined yet
func (e *Environment) Update() error {
	gotErr := false

	// If an environment variable name is specified and specified without value
	// then it inherits the value from Updatecli process environment
	if len(e.Value) == 0 && len(e.Value) > 0 {
		value, found := os.LookupEnv(e.Name)

		if !found {
			logrus.Warningf("environment variable %q not found, skipping", e.Name)
			gotErr = true
		}
		e.Value = value

	}
	if gotErr {
		return fmt.Errorf("wrong configuration")
	}
	return nil
}

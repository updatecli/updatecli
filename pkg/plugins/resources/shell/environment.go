package shell

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

const (
	// DryRunVariableName specifies the environment variable used within shell script to detect if we are in dryrun mode
	DryRunVariableName = "DRY_RUN"
)

type Environment struct {
	// Name defines the environment variable name
	Name string `yaml:",omitempty"`
	// Value defines the environment variable value
	Value string `yaml:",omitempty"`
}

func (e *Environment) String() string {
	return fmt.Sprintf("%s=%s", e.Name, e.Value)
}

func (e *Environment) Validate() error {
	gotErr := false

	// If a environment variable name specified without value
	// then inherit the value from Updatecli environment
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

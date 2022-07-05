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
	// Inherit is boolean which if the environment value is inherited from the Updatecli environment
	Inherit bool `yaml:",omitempty"`
}

func (e *Environment) String() string {
	if !e.Inherit {
		return fmt.Sprintf("%s=%s", e.Name, e.Value)
	}

	value, found := os.LookupEnv(e.Name)
	if !found {
		logrus.Warningf("environment variable %q not found, skipping", e.Name)
		return ""
	}

	return fmt.Sprintf("%s=%s", e.Name, value)
}

func (e *Environment) Validate() error {
	gotErr := false
	if e.Inherit && len(e.Value) > 0 {
		gotErr = true
	}
	if gotErr {
		return fmt.Errorf("wrong configuration")
	}
	return nil
}

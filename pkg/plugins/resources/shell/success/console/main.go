package console

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Spec is an empty struct used as a placeholder for the jsonschema.
type Spec struct {
}

type Console struct {
	exitCode *int
	output   *string
}

func New(exitCode *int, output *string) (*Console, error) {
	if exitCode == nil {
		return nil, errors.New("exitCode pointer is null")
	}
	if output == nil {
		return nil, errors.New("output pointer is null")
	}
	return &Console{
		exitCode: exitCode,
		output:   output,
	}, nil
}

// PreCommand defines operations needed to be executed before the shell command
func (c *Console) PreCommand(workingDir string) error {
	return nil
}

// PostCommand defines operations needed to be executed after the shell command
func (c *Console) PostCommand(workingDir string) error {
	return nil
}

// SourceResult defines the success criteria for a source using the shell resource
func (c *Console) SourceResult(resultSource *result.Source) error {
	switch *c.exitCode {
	case 0:
		resultSource.Information = *c.output
		resultSource.Result = result.SUCCESS
		resultSource.Description = "shell command executed successfully"

		return nil

	default:
		return fmt.Errorf("shell command failed. Expected exit code 0 but got %d", *c.exitCode)
	}
}

// ConditionResult defines the success criteria for a condition using the shell resource
func (c *Console) ConditionResult() (bool, error) {
	switch *c.exitCode {
	case 0:
		return true, nil
	default:
		logrus.Infof("shell command failed. Expected exit code 0 but got %d", *c.exitCode)
		return false, nil
	}
}

// TargetResult defines the success criteria for a target using the shell resource
func (c *Console) TargetResult() (bool, error) {
	switch *c.exitCode {
	case 0:
		if *c.output == "" {
			return false, nil
		}
		return true, nil
	default:
		return false, fmt.Errorf("shell command failed. Expected exit code 0 but got %d", *c.exitCode)
	}
}

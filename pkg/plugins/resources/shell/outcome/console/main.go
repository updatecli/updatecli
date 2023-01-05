package console

import (
	"errors"
	"fmt"
)

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

// PreCommand defines operations needed to be executed before the shell commmand
func (c *Console) PreCommand() error {
	return nil
}

// PostCommand defines operations needed to be executed after the shell command
func (c *Console) PostCommand() error {
	return nil
}

// SourceOutcome defines the success criteria for a source using the shell resource
func (c *Console) SourceResult() (string, error) {
	switch *c.exitCode {
	case 0:
		return *c.output, nil
	default:
		return "", fmt.Errorf("shell command failed. Expected exit code 0 but got %d", *c.exitCode)
	}
}

// ConditionResult defines the success criteria for a condition using the shell resource
func (c *Console) ConditionResult() (bool, error) {
	switch *c.exitCode {
	case 0:
		return true, nil
	default:
		return false, fmt.Errorf("shell command failed. Expected exit code 0 but got %d", *c.exitCode)
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

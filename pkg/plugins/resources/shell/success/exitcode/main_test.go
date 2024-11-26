package exitcode

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestSourceResult(t *testing.T) {

	dataset := []struct {
		name                       string
		exitCode                   int
		spec                       Spec
		stdout                     string
		expectedResultOutput       []result.SourceInformation
		expectedResultErrorMessage string
		expectedError              bool
		expectedNewError           bool
		expectedNewErrorMessage    error
	}{
		{
			name: "Test succeeded without command output",
			spec: Spec{
				Warning: 2,
				Success: 0,
				Failure: 1,
			},
			exitCode: 0,
			stdout:   "",
			expectedResultOutput: []result.SourceInformation{{
				Value: "",
			}},
		},
		{
			name:     "Test succeeded without specifying spec",
			exitCode: 0,
			stdout:   "",
			expectedResultOutput: []result.SourceInformation{{
				Value: "",
			}},
		},
		{
			name:     "Test failed without specifying spec",
			exitCode: 2,
			stdout:   "",
		},
		{
			name: "Test succeeded without command output",
			spec: Spec{
				Warning: 1,
				Success: 2,
				Failure: 1,
			},
			exitCode: 2,
			stdout:   "",
			expectedResultOutput: []result.SourceInformation{{
				Value: "",
			}},
		},
		{
			name:     "Test succeeded with command output",
			exitCode: 0,
			spec: Spec{
				Warning: 2,
				Success: 0,
				Failure: 1,
			},
			stdout: "1.2.3",
			expectedResultOutput: []result.SourceInformation{{
				Value: "1.2.3",
			}},
		},
		{
			name:     "Test failed with no command output",
			exitCode: 1,
			spec: Spec{
				Warning: 2,
				Success: 0,
				Failure: 1,
			},
			stdout:                     "",
			expectedResultErrorMessage: "shell command failed. Expected exit code 0 but got 1",
			expectedError:              true,
		},
		{
			name:     "Test failed with command output",
			exitCode: 2,
			spec: Spec{
				Warning: 2,
				Success: 0,
				Failure: 1,
			},
			stdout:                     "1.2.3",
			expectedResultErrorMessage: "shell command failed. Expected exit code 0 but got 2",
			expectedError:              true,
		},
		{
			name:     "Test failed with wrong exitcode combination",
			exitCode: 2,
			spec: Spec{
				Warning: 2,
				Success: 0,
				Failure: 0,
			},
			stdout:                  "1.2.3",
			expectedError:           true,
			expectedNewError:        true,
			expectedNewErrorMessage: errors.New("wrong exit spec"),
		},
	}

	for i := range dataset {
		d := dataset[i]
		t.Run(d.name, func(t *testing.T) {
			c, gotErr := New(d.spec, &d.exitCode, &d.stdout)

			switch d.expectedNewError {
			case true:
				assert.Equal(t, gotErr, d.expectedNewErrorMessage)
				return
			case false:
				assert.NoError(t, gotErr)
			}

			gotResult := result.Source{}
			gotErr = c.SourceResult(&gotResult)

			assert.Equal(t, d.expectedResultOutput, gotResult.Information)
			switch d.expectedError {
			case true:
				assert.Equal(t, gotErr.Error(), d.expectedResultErrorMessage)
			}
		})
	}
}

func TestConditionResult(t *testing.T) {

	dataset := []struct {
		name                       string
		exitCode                   int
		spec                       Spec
		stdout                     string
		expectedResultOutput       bool
		expectedResultErrorMessage error
		expectedError              bool
		expectedNewError           bool
		expectedNewErrorMessage    error
	}{
		{
			name:                 "Test succeeded with no command output",
			exitCode:             0,
			stdout:               "",
			expectedResultOutput: true,
			spec: Spec{
				Warning: 2,
				Success: 0,
				Failure: 1,
			},
		},
		{
			name:                 "Test succeeded with command output",
			exitCode:             0,
			stdout:               "1.2.3",
			expectedResultOutput: true,
			spec: Spec{
				Warning: 2,
				Success: 0,
				Failure: 1,
			},
		},
		{
			name:                 "Test succeeded with command output",
			exitCode:             1,
			stdout:               "1.2.3",
			expectedResultOutput: true,
			spec: Spec{
				Warning: 2,
				Success: 1,
				Failure: 0,
			},
		},
		{
			name:                       "Test failed with no command output",
			exitCode:                   1,
			stdout:                     "",
			expectedResultOutput:       false,
			expectedResultErrorMessage: errors.New("shell command failed. Expected exit code 0 but got 1"),
			expectedError:              true,
			spec: Spec{
				Warning: 2,
				Success: 0,
				Failure: 1,
			},
		},
		{
			name:                       "Test failed with command output",
			exitCode:                   2,
			stdout:                     "1.2.3",
			expectedResultOutput:       false,
			expectedResultErrorMessage: errors.New("shell command failed. Expected exit code 0 but got 2"),
			expectedError:              true,
			spec: Spec{
				Warning: 2,
				Success: 0,
				Failure: 1,
			},
		},
	}

	for i := range dataset {
		d := dataset[i]
		t.Run(d.name, func(t *testing.T) {
			c, gotErr := New(d.spec, &d.exitCode, &d.stdout)

			switch d.expectedNewError {
			case true:
				assert.Equal(t, gotErr, d.expectedNewErrorMessage)
				return
			case false:
				assert.NoError(t, gotErr)
			}

			gotConditionResult, gotErr := c.ConditionResult()

			assert.Equal(t, gotConditionResult, d.expectedResultOutput)
			switch d.expectedError {
			case true:
				assert.Equal(t, gotErr, d.expectedResultErrorMessage)
			case false:
				assert.NoError(t, gotErr)
			}
		})
	}
}

func TestTargetResult(t *testing.T) {

	dataset := []struct {
		name                       string
		spec                       Spec
		exitCode                   int
		stdout                     string
		expectedResultOutput       bool
		expectedResultErrorMessage error
		expectedError              bool
		expectedNewError           bool
		expectedNewErrorMessage    error
	}{
		{
			name:                 "Test succeeded exit code 0",
			exitCode:             0,
			expectedResultOutput: false,
			spec: Spec{
				Warning: 2,
				Success: 0,
				Failure: 1,
			},
		},
		{
			name:                 "Triggered changed with command output",
			exitCode:             2,
			expectedResultOutput: true,
			spec: Spec{
				Warning: 2,
				Success: 0,
				Failure: 1,
			},
		},
		{
			name:                       "Test failed with no command output",
			exitCode:                   1,
			expectedResultOutput:       false,
			expectedResultErrorMessage: errors.New("shell command failed. Expected exit code 0 but got 1"),
			expectedError:              true,
			spec: Spec{
				Warning: 2,
				Success: 0,
				Failure: 1,
			},
		},
		{
			name:                       "Test failed with command output",
			exitCode:                   1,
			stdout:                     "1.2.3",
			expectedResultOutput:       false,
			expectedResultErrorMessage: errors.New("shell command failed. Expected exit code 0 but got 1"),
			expectedError:              true,
			spec: Spec{
				Warning: 2,
				Success: 0,
				Failure: 1,
			},
		},
	}

	for i := range dataset {
		d := dataset[i]
		t.Run(d.name, func(t *testing.T) {
			c, gotErr := New(d.spec, &d.exitCode, &d.stdout)

			switch d.expectedNewError {
			case true:
				assert.Equal(t, gotErr, d.expectedNewErrorMessage)
				return
			case false:
				assert.NoError(t, gotErr)
			}

			gotTargetResult, gotErr := c.TargetResult()

			assert.Equal(t, gotTargetResult, d.expectedResultOutput)
			switch d.expectedError {
			case true:
				assert.Equal(t, gotErr, d.expectedResultErrorMessage)
			case false:
				assert.NoError(t, gotErr)
			}
		})
	}
}

func TestPreCommand(t *testing.T) {
	exitCode := 3
	stdout := "1.2.3"
	spec := Spec{
		Warning: 2,
		Success: 0,
		Failure: 1,
	}

	c, gotErr := New(spec, &exitCode, &stdout)
	assert.NoError(t, gotErr)

	if c.PreCommand("") != nil {
		t.Fail()
	}
}

func TestPostCommand(t *testing.T) {
	exitCode := 3
	stdout := "1.2.3"
	spec := Spec{
		Warning: 2,
		Success: 0,
		Failure: 1,
	}

	c, gotErr := New(spec, &exitCode, &stdout)
	assert.NoError(t, gotErr)

	if c.PostCommand("") != nil {
		t.Fail()
	}
}

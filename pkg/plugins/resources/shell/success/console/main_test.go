package console

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
		stdout                     string
		expectedResultOutput       []result.SourceInformation
		expectedResultErrorMessage error
		expectedError              bool
		expectedNewError           bool
		expectedNewErrorMessage    error
	}{
		{
			name:     "Test succeeded without command output",
			exitCode: 0,
			stdout:   "",
			expectedResultOutput: []result.SourceInformation{{
				Value: "",
			}},
		},
		{
			name:     "Test succeeded with command output",
			exitCode: 0,
			stdout:   "1.2.3",
			expectedResultOutput: []result.SourceInformation{{
				Value: "1.2.3",
			}},
		},
		{
			name:          "Test failed with no command output",
			exitCode:      1,
			stdout:        "",
			expectedError: false,
		},
		{
			name:     "Test failed with command output",
			exitCode: 2,
			stdout:   "1.2.3",
		},
	}

	for i := range dataset {
		t.Run(dataset[i].name, func(t *testing.T) {
			c, gotErr := New(&dataset[i].exitCode, &dataset[i].stdout)
			switch dataset[i].expectedNewError {
			case true:
				assert.Equal(t, gotErr, dataset[i].expectedNewErrorMessage)
				return
			case false:
				assert.NoError(t, gotErr)
			}

			gotResult := result.Source{}
			gotErr = c.SourceResult(&gotResult)

			assert.Equal(t, dataset[i].expectedResultOutput, gotResult.Information)
			switch dataset[i].expectedError {
			case true:
				assert.Equal(t, gotErr, dataset[i].expectedResultErrorMessage)
			}

		})
	}
}

func TestConditionResult(t *testing.T) {

	dataset := []struct {
		name                       string
		exitCode                   int
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
		},
		{
			name:                 "Test succeeded with command output",
			exitCode:             0,
			stdout:               "1.2.3",
			expectedResultOutput: true,
		},
		{
			name:                 "Test failed with no command output",
			exitCode:             1,
			stdout:               "",
			expectedResultOutput: false,
		},
		{
			name:                 "Test failed with command output",
			exitCode:             2,
			stdout:               "1.2.3",
			expectedResultOutput: false,
		},
	}

	for i := range dataset {
		d := dataset[i]

		t.Run(d.name, func(t *testing.T) {
			c, gotErr := New(&d.exitCode, &d.stdout)

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
		exitCode                   int
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
			expectedResultOutput: false,
		},
		{
			name:                 "Test succeeded with command output",
			exitCode:             0,
			stdout:               "1.2.3",
			expectedResultOutput: true,
		},
		{
			name:                       "Test failed with no command output",
			exitCode:                   1,
			stdout:                     "",
			expectedResultOutput:       false,
			expectedResultErrorMessage: errors.New("shell command failed. Expected exit code 0 but got 1"),
			expectedError:              true,
		},
		{
			name:                       "Test failed with command output",
			exitCode:                   2,
			stdout:                     "1.2.3",
			expectedResultOutput:       false,
			expectedResultErrorMessage: errors.New("shell command failed. Expected exit code 0 but got 2"),
			expectedError:              true,
		},
	}

	for i := range dataset {
		d := dataset[i]
		t.Run(d.name, func(t *testing.T) {
			c, gotErr := New(&d.exitCode, &d.stdout)

			switch d.expectedNewError {
			case true:
				assert.Equal(t, gotErr, d.expectedNewErrorMessage)
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

	c, gotErr := New(&exitCode, &stdout)
	assert.NoError(t, gotErr)

	if c.PreCommand("") != nil {
		t.Fail()
	}
}

func TestPostCommand(t *testing.T) {
	exitCode := 3
	stdout := "1.2.3"

	c, gotErr := New(&exitCode, &stdout)
	assert.NoError(t, gotErr)

	if c.PostCommand("") != nil {
		t.Fail()
	}
}

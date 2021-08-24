package shell

import (
	"github.com/updatecli/updatecli/pkg/core/scm"
)

// mockCommandExecutor is a stub implementation of the `commandExecutor` interface
// to be used in our test suite.
// It stores the received `command` and returns the preconfigured `result` and `err`.
type mockCommandExecutor struct {
	gotCommand command
	result     commandResult
	err        error
}

func (mce *mockCommandExecutor) ExecuteCommand(cmd command) (commandResult, error) {
	mce.gotCommand = cmd
	return mce.result, mce.err
}

// mocking SCM object (no introspection: only get values)
type mockScm struct {
	scm.Scm

	workingDir string
}

func (m *mockScm) GetDirectory() (directory string) {
	return m.workingDir
}

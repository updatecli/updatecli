package shell

import "fmt"

////////////// Test Utilities method aimed at mocking calls to other objects

// mocking commandExecutor.ExecuteCommand
type mockCommandExecutor struct {
	gotCommand command
	result     commandResult
	err        error
}

func (mce *mockCommandExecutor) ExecuteCommand(cmd command) (commandResult, error) {
	mce.gotCommand = cmd
	if mce.gotCommand.Cmd == "" {
		return commandResult{}, fmt.Errorf(ErrEmptyCommand)
	}
	return mce.result, mce.err
}

// mocking SCM object (no introspection: only get values)
type mockScm struct {
	workingDir string
}

func (m *mockScm) Add(files []string) error {
	return nil
}
func (m *mockScm) Clone() (string, error) {
	return "", nil
}
func (m *mockScm) Checkout() error {
	return nil
}
func (m *mockScm) GetDirectory() (directory string) {
	return m.workingDir
}
func (m *mockScm) Init(source string, pipelineID string) error {
	return nil
}
func (m *mockScm) Push() error {
	return nil
}
func (m *mockScm) Commit(message string) error {
	return nil
}
func (m *mockScm) Clean() error {
	return nil
}
func (m *mockScm) PushTag(tag string) error {
	return nil
}

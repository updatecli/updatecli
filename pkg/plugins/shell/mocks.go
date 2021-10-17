package shell

// MockCommandExecutor is a stub implementation of the `commandExecutor` interface
// to be used in our test suite.
// It stores the received `command` and returns the preconfigured `result` and `err`.
type MockCommandExecutor struct {
	GotCommand command
	Result     commandResult
	Err        error
}

func (mce *MockCommandExecutor) ExecuteCommand(cmd command) (commandResult, error) {
	mce.GotCommand = cmd
	return mce.Result, mce.Err
}

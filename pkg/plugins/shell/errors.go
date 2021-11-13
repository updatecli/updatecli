package shell

type ErrEmptyCommand struct{}

func (e *ErrEmptyCommand) Error() string {
	return "Invalid spec for shell resource: command is empty."
}

type ExecutionFailedError struct{}

func (e *ExecutionFailedError) Error() string {
	return "Shell command exited on error."
}

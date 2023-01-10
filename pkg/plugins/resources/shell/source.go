package shell

// Source returns the stdout of the shell command if its exit code is 0
// otherwise an error is returned with the content of stderr
func (s *Shell) Source(workingDir string) (string, error) {

	// Ensure environment variable(s) are up to date
	// either it already has a value specified, or it retrieves
	// the value from the Updatecli process
	err := s.spec.Environments.Load()
	if err != nil {
		return "", &ExecutionFailedError{}
	}

	// Provides the "UPDATECLI_PIPELINE_STAGE" environment variable set to "source"
	// It's only purpose is to have at least one environment variable
	// so we don't fallback to use the current process environment as explained
	// on https://pkg.go.dev/os/exec#Cmd
	// Provides the "DRY_RUN" environment variable to the shell command (true if "diff", false if "apply")
	env := append(s.spec.Environments, Environment{
		Name:  CurrentStageVariableName,
		Value: "source",
	})

	// PreCommand is executed to collect information before running the shell command
	// so we could collect information needed to validate that a command successfully as expected
	err = s.success.PreCommand()
	if err != nil {
		return "", err
	}

	s.executeCommand(command{
		Cmd: s.spec.Command,
		Dir: workingDir,
		Env: env.ToStringSlice(),
	})

	// PostCommand is executed to collect information after running the shell command
	// so we could collect information needed to validate that a command successfully as expected
	err = s.success.PostCommand()
	if err != nil {
		return "", err
	}

	return s.success.SourceResult()
}

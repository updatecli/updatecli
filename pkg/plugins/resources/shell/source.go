package shell

// Source returns the stdout of the shell command if its exit code is 0
// otherwise an error is returned with the content of stderr
func (s *Shell) Source(workingDir string) (string, error) {

	s.executeCommand(command{
		Cmd: s.spec.Command,
		Dir: workingDir,
		Env: s.spec.Environments.ToStringArray(),
	})

	if s.result.ExitCode != 0 {
		return "", &ExecutionFailedError{}
	}

	return s.result.Stdout, nil
}

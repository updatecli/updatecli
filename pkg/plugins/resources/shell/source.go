package shell

import "fmt"

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

	scriptFilename, err := newShellScript(s.spec.Command)
	if err != nil {
		return "", fmt.Errorf("failed initializing source script - %s", err)
	}

	s.executeCommand(command{
		Cmd: s.interpreter + " " + scriptFilename,
		Dir: workingDir,
		Env: env.ToStringSlice(),
	})

	if s.result.ExitCode != 0 {
		return "", &ExecutionFailedError{}
	}

	return s.result.Stdout, nil
}

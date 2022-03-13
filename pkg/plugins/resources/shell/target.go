package shell

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

// Target executes the provided command (concatenated with the source) to apply the change.
//	The command is expected, if it changes something, to print the new value to the stdout
//	- An exit code of 0 and an empty stdout means: "successful command and no change"
//	- An exit code of 0 and something on the stdout means: "successful command with a changed value"
//	- Any other exit code means "failed command with no change"
// The environment variable 'DRY_RUN' is set to true or false based on the input parameter (e.g. 'updatecli diff' or 'apply'?)
func (s *Shell) Target(source, workingDir string, dryRun bool) (bool, []string, string, error) {

	// Ensure environment variable(s) are up to date
	// either it already has a value specified, or it retrieves
	// the value from the Updatecli process
	err := s.spec.Environments.Load()
	if err != nil {
		return false, []string{}, "", nil
	}

	// Provides the "UPDATECLI_PIPELINE_STAGE" environment variable set to "target"
	// It's only purpose is to have at least one environment variable
	// so we don't fallback to use the current process environment as explained
	// on https://pkg.go.dev/os/exec#Cmd
	env := append(s.spec.Environments, Environment{
		Name:  CurrentStageVariableName,
		Value: "target",
	})

	// Provides the "DRY_RUN" environment variable to the shell command (true if "diff", false if "apply")
	env = append(env, Environment{
		Name:  DryRunVariableName,
		Value: fmt.Sprintf("%v", dryRun),
	})

	s.executeCommand(command{
		Cmd: s.appendSource(source),
		Dir: workingDir,
		Env: env.ToStringSlice(),
	})

	if s.result.ExitCode != 0 {
		return false, []string{}, "", &ExecutionFailedError{}
	}

	if s.result.Stdout == "" {
		logrus.Info("No change detected")
		return false, []string{}, "", nil
	}

	// Retrieve the list of modified files using git (could be improved by not assuming git but this logic should be moved as part of target instead of only here)
	nativeGitHandler := gitgeneric.GoGit{}
	files, _ := nativeGitHandler.GetChangedFiles(workingDir)
	commitMessage := fmt.Sprintf("ran shell command %q which changed the files %s", s.appendSource(source), files)

	return true, files, commitMessage, nil
}

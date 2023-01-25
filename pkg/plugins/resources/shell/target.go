package shell

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func (s *Shell) Target(source string, dryRun bool) (bool, error) {
	changed, _, err := s.target(source, "", dryRun)
	return changed, err
}

func (s *Shell) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (bool, []string, string, error) {
	changed, message, err := s.target(source, scm.GetDirectory(), dryRun)
	if err != nil {
		return false, []string{}, "", err
	}

	// Once the changes have been applied inside the scm's temp directory, then we have to get the list of these changes
	files, err := scm.GetChangedFiles(scm.GetDirectory())
	if err != nil {
		return false, []string{}, "", err
	}
	return changed, files, message, err
}

// Target executes the provided command (concatenated with the source) to apply the change.
// The command is expected, if it changes something, to print the new value to the stdout
//   - An exit code of 0 and an empty stdout means: "successful command and no change"
//   - An exit code of 0 and something on the stdout means: "successful command with a changed value"
//   - Any other exit code means "failed command with no change"
//
// The environment variable 'DRY_RUN' is set to true or false based on the input parameter (e.g. 'updatecli diff' or 'apply'?)
func (s *Shell) target(source, workingDir string, dryRun bool) (bool, string, error) {

	// Ensure environment variable(s) are up to date
	// either it already has a value specified, or it retrieves
	// the value from the Updatecli process
	err := s.spec.Environments.Load()
	if err != nil {
		return false, "", nil
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

	scriptFilename, err := newShellScript(s.appendSource(source))
	if err != nil {
		return false, "", fmt.Errorf("failed initializing source script - %s", err)
	}

	err = s.executeCommand(command{
		Cmd: s.interpreter + " " + scriptFilename,
		Dir: s.getWorkingDirPath(workingDir),
		Env: env.ToStringSlice(),
	})
	if err != nil {
		return false, "", fmt.Errorf("failed while running target script - %s", err)
	}

	if s.result.ExitCode != 0 {
		return false, "", &ExecutionFailedError{}
	}

	if s.result.Stdout == "" {
		logrus.Info("No change detected")
		return false, "", nil
	}

	commitMessage := fmt.Sprintf("ran shell command %q", s.appendSource(source))

	return true, commitMessage, nil
}

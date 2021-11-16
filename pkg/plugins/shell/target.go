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
//	The command is expected, if it changes something, to print the new value to the stdout
//	- An exit code of 0 and an empty stdout means: "successful command and no change"
//	- An exit code of 0 and something on the stdout means: "successful command with a changed value"
//	- Any other exit code means "failed command with no change"
// The environment variable 'DRY_RUN' is set to true or false based on the input parameter (e.g. 'updatecli diff' or 'apply'?)
func (s *Shell) target(source, workingDir string, dryRun bool) (bool, string, error) {
	s.executeCommand(command{
		Cmd: s.appendSource(source),
		Dir: workingDir,
		// Provides the "DRY_RUN" environment variable to the shell command (true if "diff", false if "apply")
		Env: []string{fmt.Sprintf("DRY_RUN=%v", dryRun)},
	})

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

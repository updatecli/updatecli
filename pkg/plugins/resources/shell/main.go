package shell

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

// Spec defines a specification for a "shell" resource
// parsed from an updatecli manifest file
type Spec struct {
	// command specifies the shell command to execute by Updatecli
	Command string `yaml:",omitempty" jsonschema:"required"`
	// environments allows to pass environment variable(s) to the shell script. By default no environment variable are shared.
	Environments Environments `yaml:",omitempty"`
	// ChangedIf defines how to interpreted shell command success criteria. What a success means, what an error means, and what a warning would mean
	ChangedIf SpecChangedIf `yaml:",omitempty" json:",omitempty"`
	// Shell specifies which shell interpreter to use. Default to powershell(Windows) and "/bin/sh" (Darwin/Linux)
	Shell string `yaml:",omitempty"`
	// workdir specifies the working directory path from where to execute the command. It defaults to the current context path (scm or current shell). Updatecli join the current path and the one specified in parameter if the parameter one contains a relative path.
	WorkDir string `yaml:",omitempty"`
}

// Shell defines a resource of kind "shell"
type Shell struct {
	executor    commandExecutor
	spec        Spec
	result      commandResult
	success     Successer
	interpreter string
}

// New returns a reference to a newly initialized Shell object from a ShellSpec
// or an error if the provided ShellSpec triggers a validation error.
func New(spec interface{}) (*Shell, error) {
	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	if newSpec.Command == "" {
		return nil, &ErrEmptyCommand{}
	}

	err = newSpec.Environments.Validate()
	if err != nil {
		return nil, err
	}

	interpreter := getDefaultShell()
	if newSpec.Shell != "" {
		interpreter = newSpec.Shell
	}

	s := Shell{
		executor:    &nativeCommandExecutor{},
		spec:        newSpec,
		interpreter: interpreter,
	}

	err = s.InitChangedIf()
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func getDefaultShell() string {
	os := runtime.GOOS

	switch os {
	case WINOS:
		// pwshell is the default shell on Windows system
		return "powershell.exe -executionpolicy remotesigned -File"
	default:
		return "/bin/sh"
	}

}

// appendSource appends the source as last argument if not empty.
func (s *Shell) appendSource(source string) string {
	// Append the source as last argument if not empty
	if source != "" {
		return s.spec.Command + " " + source
	}

	return s.spec.Command
}

// executeCommand call the shell command executor to execute its command
// and sets the internal "result" to the command result
func (s *Shell) executeCommand(inputCmd command) (err error) {

	s.result, err = s.executor.ExecuteCommand(inputCmd)
	// Logs the result
	s.report()

	return err
}

// report logs the result of the shell command to the end user.
func (s *Shell) report() {
	message := fmt.Sprintf("The shell 🐚 command %q", s.result.Cmd)
	stdoutMessage := fmt.Sprintf("with the following output:\n%s", formatShellBlock(s.result.Stdout))
	stderrMessage := fmt.Sprintf("command stderr output was:\n%s", formatShellBlock(s.result.Stderr))

	if s.result.ExitCode != 0 {
		// Shell command exited with an error: log everything as info, including exit code and stderr
		message += fmt.Sprintf(" exited on error (exit code %d) %s\n\n%s", s.result.ExitCode, stdoutMessage, stderrMessage)

		logrus.Info(message)
		return
	}

	// Shell command ran successfully: logs the command and its standard output as info, and stderr as debug
	message += fmt.Sprintf(" ran successfully %s", stdoutMessage)

	logrus.Info(message)
	logrus.Debug(stderrMessage)
}

func formatShellBlock(content string) string {
	const logShellBlockSeparator string = "----"
	message := fmt.Sprintf("%s\n", logShellBlockSeparator)

	if content != "" {
		message += fmt.Sprintf("%s\n", content)
	}

	message += logShellBlockSeparator

	return message
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (s *Shell) Changelog() string {
	return ""
}

// getWorkingDirPath returns the real workingDir path that should be used by the shell resource
func (s *Shell) getWorkingDirPath(currentWorkDir string) string {
	if s.spec.WorkDir == "" {
		return currentWorkDir
	}

	if filepath.IsAbs(s.spec.WorkDir) {
		return s.spec.WorkDir
	}

	return filepath.Join(currentWorkDir, s.spec.WorkDir)
}

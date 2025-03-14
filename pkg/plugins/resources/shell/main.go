package shell

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Spec defines a specification for a "shell" resource
// parsed from an updatecli manifest file
type Spec struct {
	// command specifies the shell command to execute by Updatecli
	//
	// default:
	//   empty
	//
	// remark:
	//   When the shell plugin is used in the context of condition, or target, the default source output is passed as an argument to the shell command.
	//   for example the two following snippets are equivalent:
	//
	//   ---
	//   targets:
	//     default:
	//       name: Example 2
	//       kind: shell
	//       sourceid: default
	//       spec:
	//         command: 'echo'
	//   ---
	//   targets:
	//     default:
	//       name: Example 2
	//       kind: shell
	//       disablesourceinput: true
	//       spec:
	//         command: 'echo {{ source "default"}}'
	//   ---

	Command string `yaml:",omitempty" jsonschema:"required"`
	// environments allows to pass environment variable(s) to the shell script.
	//
	//  default:
	//     If environments is unset then it depends on the operating system.
	//       - Windows: ["PATH","", "PSModulePath", "PSModuleAnalysisCachePath", "", "PATHEXT", "", "TEMP", "", "HOME", "", "USERPROFILE", "", "PROFILE"]
	//       - Darwin/Linux: ["PATH", "", "HOME", "", "USER", "", "LOGNAME", "", "SHELL", "", "LANG", "", "LC_ALL"]
	//
	// remark:
	//   For security reason, Updatecli doesn't pass the entire environment to the shell command but instead works
	//   with an allow list of environment variables.
	//
	Environments *Environments `yaml:",omitempty"`
	// ChangedIf defines how to interprete shell command execution.
	// What a success means, what an error means, and what a warning would mean in the context of Updatecli.
	//
	// Please note that in the context of Updatecli,
	//  - a success means nothing changed
	//  - a warning means something changed
	//  - an error means something went wrong
	//
	// Changedif can be of kind "exitcode", "console/output", or "file/checksum"
	//
	//   "console/output" (default)
	//     Check the output of the command to identify if Updatecli should report a success, a warning, or an error.
	//     If a target returns anything to stdout, Updatecli interprets it as a something changed, otherwise it's a success.
	//
	//     example:
	//
	//
	//     ---
	//     targets:
	//       default:
	//         name: 'doc: synchronize release note'
	//         kind: 'shell'
	//         disablesourceinput: true
	//         spec:
	//           command: 'releasepost --dry-run="$DRY_RUN" --config {{ .config }} --clean'
	//     ---
	//
	//   "exitcode":
	//     Check the exit code of the command to identify if Updatecli should report a success, a warning, or an error.
	//
	//     example:
	//
	//     ---
	//     targets:
	//       default:
	//         name: 'doc: synchronize release note'
	//         kind: 'shell'
	//         disablesourceinput: true
	//         spec:
	//           command: 'releasepost --dry-run="$DRY_RUN" --config {{ .config }} --clean'
	//           environments:
	//             - name: 'GITHUB_TOKEN'
	//             - name: 'PATH'
	//           changedif:
	//             kind: 'exitcode'
	//             spec:
	//               warning: 0
	//               success: 1
	//               failure: 2
	//     ---
	//
	//
	//   "file/checksum":
	//     Check the checksum of file(s) to identify if Updatecli should report a success, a warning, or an error.
	//
	//     example:
	//
	//     ---
	//     targets:
	//       default:
	//         disablesourceinput: true
	//         name: Example of a shell command with a checksum success criteria
	//         kind: shell
	//         spec:
	//           command: |
	//     	  	   yq -i '.a.b[0].c = "cool"' file.yaml
	//           changedif:
	//             kind: file/checksum
	//             spec:
	//               files:
	//                 - file.yaml
	//     ---
	//
	//
	//
	ChangedIf SpecChangedIf `yaml:",omitempty" json:",omitempty"`
	// Shell specifies which shell interpreter to use.
	//
	// default:
	//   Depends on the operating system:
	//     - Windows: "powershell"
	//     - Darwin/Linux: "/bin/sh"
	//
	Shell string `yaml:",omitempty"`
	// workdir specifies the working directory path from where to execute the command. It defaults to the current context path (scm or current shell). Updatecli join the current path and the one specified in parameter if the parameter one contains a relative path.
	//
	// default: If a scmid is specified then the default
	WorkDir string `yaml:",omitempty"`
}

// Shell defines a resource of kind "shell"
type Shell struct {
	executor     commandExecutor
	spec         Spec
	result       commandResult
	success      Successer
	interpreter  string
	environments Environments
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

	interpreter := getDefaultShell()
	if newSpec.Shell != "" {
		interpreter = newSpec.Shell
	}

	environments := Environments{}

	if newSpec.Environments != nil {
		environments = *newSpec.Environments
	} else {
		switch runtime.GOOS {
		case WINOS:
			environments = DefaultWinEnvVariables
		default:
			environments = DefaultUnixEnvVariables
		}
	}

	err = environments.Validate()
	if err != nil {
		return nil, err
	}

	s := Shell{
		executor:     &nativeCommandExecutor{},
		spec:         newSpec,
		interpreter:  interpreter,
		environments: environments,
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
func (s *Shell) Changelog(from, to string) *result.Changelogs {
	return nil
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

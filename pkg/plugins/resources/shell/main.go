package shell

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

/*
"shell" defines the specification for running a shell command.
It can be used as a "source", a "condition", or a "target".
*/
type Spec struct {
	// "command" defines the shell command to run.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * "command" is required.
	//   * in a condition or a target, the source output is appended to the command as its last argument.
	//     The two following snippets are equivalent:
	//
	// ```
	//   targets:
	//     default:
	//       name: Example 2
	//       kind: shell
	//       sourceid: default
	//       spec:
	//         command: 'echo'
	// ```
	//
	// ```
	//   targets:
	//     default:
	//       name: Example 2
	//       kind: shell
	//       disablesourceinput: true
	//       spec:
	//         command: 'echo {{ source "default"}}'
	// ```
	//
	Command string `yaml:",omitempty" jsonschema:"required"`
	// "environments" defines the environment variables passed to the shell command.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// default:
	//   depends on the operating system:
	//   * Windows: PATH, PSModulePath, PSModuleAnalysisCachePath, PATHEXT, TEMP, HOME, USERPROFILE, PROFILE
	//   * Darwin/Linux: PATH, HOME, USER, LOGNAME, SHELL, LANG, LC_ALL
	//
	// remark:
	//   * for security reasons, Updatecli does not pass its whole environment to the shell command.
	//     It uses an allow list of environment variables instead.
	//   * "DRY_RUN" is reserved and set by Updatecli, so it cannot be defined.
	//   * "UPDATECLI_PIPELINE_STAGE" is set by Updatecli to the current stage.
	//
	// example:
	//   * environments:
	//       - name: PATH
	//       - name: GITHUB_TOKEN
	//
	Environments *Environments `yaml:",omitempty"`
	// "changedif" defines how Updatecli interprets the result of the shell command.
	//
	// In Updatecli, a success means nothing changed, a warning means something
	// changed, and an error means something went wrong.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// default:
	//   console/output
	//
	// remark:
	//   * accepted kinds are "console/output", "exitcode" and "file/checksum".
	//   * "console/output" checks the command output. In a target, any output on
	//     stdout means something changed, otherwise nothing changed.
	//   * "exitcode" checks the command exit code.
	//   * "file/checksum" checks the checksum of files before and after the command.
	//
	// example:
	//
	// ```
	//   targets:
	//     default:
	//       name: 'doc: synchronise release note'
	//       kind: 'shell'
	//       disablesourceinput: true
	//       spec:
	//         command: 'releasepost --dry-run="$DRY_RUN" --config {{ .config }} --clean'
	//         environments:
	//           - name: 'GITHUB_TOKEN'
	//           - name: 'PATH'
	//         changedif:
	//           kind: 'exitcode'
	//           spec:
	//             warning: 0
	//             success: 1
	//             failure: 2
	// ```
	//
	// ```
	//   targets:
	//     default:
	//       disablesourceinput: true
	//       name: Example of a shell command with a checksum success criteria
	//       kind: shell
	//       spec:
	//         command: |
	//           yq -i '.a.b[0].c = "cool"' file.yaml
	//         changedif:
	//           kind: file/checksum
	//           spec:
	//             files:
	//               - file.yaml
	// ```
	//
	ChangedIf SpecChangedIf `yaml:",omitempty" json:",omitempty"`
	// "shell" defines the shell interpreter used to run the command.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// default:
	//   depends on the operating system:
	//   * Windows: "powershell.exe -executionpolicy remotesigned -File"
	//   * Darwin/Linux: "/bin/sh"
	//
	Shell string `yaml:",omitempty"`
	// "workdir" defines the working directory from where the command runs.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// default:
	//   the scm checkout directory, or the manifest directory when no scm is set.
	//
	// remark:
	//   * a relative path is joined to the default directory.
	//   * an absolute path is used as is.
	//
	WorkDir string `yaml:",omitempty"`
}

// Shell defines a resource of kind "shell"
type Shell struct {
	executor     commandExecutor
	spec         Spec
	result       commandResult
	success      Evaluator
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

	var environments Environments
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

// getWorkingDirPath returns the directory the shell command must run from.
//
// spec.workdir is a location rather than a file Updatecli reads or writes, so it is
// resolved against the base directory without being held inside the SCM boundary.
func (s *Shell) getWorkingDirPath(pathResolver pathresolver.Resolver) string {
	if s.spec.WorkDir == "" {
		return pathResolver.Dir()
	}

	if filepath.IsAbs(s.spec.WorkDir) {
		return s.spec.WorkDir
	}

	return filepath.Join(pathResolver.Dir(), s.spec.WorkDir)
}

// ReportConfig returns a new configuration object with only the necessary fields
// to identify the resource without any sensitive information or context specific data.
func (s *Shell) ReportConfig() interface{} {
	return Spec{
		Command:   s.spec.Command,
		ChangedIf: s.spec.ChangedIf,
	}
}

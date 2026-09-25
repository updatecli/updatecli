package exitcode

import (
	"errors"
	"fmt"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Spec defines the parameters of the "exitcode" success criteria.
type Spec struct {
	// "warning" defines the exit code meaning that something changed.
	//
	// default:
	//   2
	//
	// remark:
	//   * only used in a target.
	//   * the defaults apply only when "warning", "success" and "failure" are all unset or set to 0.
	//
	Warning int `yaml:",omitempty" jsonschema:"required"`
	// "success" defines the exit code meaning that nothing changed.
	//
	// default:
	//   0
	//
	// remark:
	//   * the defaults apply only when "warning", "success" and "failure" are all unset or set to 0.
	//   * "success" and "failure" must differ.
	//
	Success int `yaml:",omitempty" jsonschema:"required"`
	// "failure" defines the exit code meaning that something went wrong.
	//
	// default:
	//   1
	//
	// remark:
	//   * the defaults apply only when "warning", "success" and "failure" are all unset or set to 0.
	//   * "success" and "failure" must differ.
	//
	Failure int `yaml:",omitempty" jsonschema:"required"`
}

type ExitCode struct {
	output   *string
	exitCode *int
	spec     Spec
}

func New(spec interface{}, exitCode *int, output *string) (*ExitCode, error) {
	var s Spec
	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return nil, err
	}

	// Updatecli assumes that if all exitcode have the default value set to 0
	// Then it means that no user input were provided and we fallback to default values
	// Where
	// success is 0
	// failure is 1
	// warning is 2
	if s.Failure == 0 && s.Warning == 0 && s.Success == 0 {
		s.Success = 0
		s.Failure = 1
		s.Warning = 2
	}
	err = s.Validate()
	if err != nil {
		return nil, err
	}

	if exitCode == nil {
		return nil, errors.New("exitCode pointer is not set")
	}
	return &ExitCode{
		exitCode: exitCode,
		output:   output,
		spec:     s,
	}, nil
}

func (s Spec) Validate() error {
	var errs []error
	if s.Failure == s.Success {
		errs = append(errs, fmt.Errorf("exit code can't be the same for success and failure - %d", s.Failure))
	}
	if len(errs) > 0 {
		for i := range errs {
			logrus.Errorln(errs[i])
		}
		return fmt.Errorf("wrong exit spec")
	}
	return nil
}

// PreCommand defines operations needed to be executed before the shell command
func (e *ExitCode) PreCommand(workingDir string) error {
	return nil
}

// PostCommand defines operations needed to be executed after the shell command
func (e *ExitCode) PostCommand(workingDir string) error {
	return nil
}

// SourceResult defines the success criteria for a source using the shell resource
func (e *ExitCode) SourceResult(resultSource *result.Source) error {
	switch *e.exitCode {
	case e.spec.Success:
		resultSource.Information = *e.output
		resultSource.Result = result.SUCCESS
		resultSource.Description = "shell command successfully executed"

		return nil

	default:
		return fmt.Errorf("shell command failed. Expected exit code %d but got %d", e.spec.Success, *e.exitCode)
	}
}

// ConditionResult defines the success criteria for a condition using the shell resource
func (e *ExitCode) ConditionResult() (bool, error) {
	switch *e.exitCode {
	case e.spec.Success:
		return true, nil
	default:
		return false, fmt.Errorf("shell command failed. Expected exit code %d but got %d", e.spec.Success, *e.exitCode)
	}
}

// TargetResult defines the success criteria for a target using the shell resource
func (e *ExitCode) TargetResult() (bool, error) {
	switch *e.exitCode {
	case e.spec.Success:
		return false, nil
	case e.spec.Warning:
		return true, nil
	case e.spec.Failure:
		return false, fmt.Errorf("shell command failed. Expected exit code %d but got %d", e.spec.Success, *e.exitCode)
	default:
		return false, fmt.Errorf("shell command failed. Expected exit code %d to succeed but got %d", e.spec.Success, *e.exitCode)
	}
}

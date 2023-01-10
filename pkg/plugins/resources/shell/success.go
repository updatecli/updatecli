package shell

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/shell/success/checksum"
	"github.com/updatecli/updatecli/pkg/plugins/resources/shell/success/console"
	"github.com/updatecli/updatecli/pkg/plugins/resources/shell/success/exitcode"
)

/*
	Result is the package containing the logic used by the Shell resource resource to identify
	if a shell command result should be considered a success, a failure, or a warning
*/

var (
	MappingSpecSuccess = map[string]interface{}{
		"console/output": nil,
		"exitcode":       exitcode.Spec{},
		"file/checksum":  checksum.Spec{},
	}
)

type SpecSuccess struct {
	Kind string                 `yaml:",omitempty"`
	Spec map[string]interface{} `yaml:",omitempty"`
}

type Successer interface {
	PreCommand() error
	PostCommand() error
	SourceResult() (string, error)
	ConditionResult() (bool, error)
	TargetResult() (bool, error)
}

func (s *Shell) InitSuccess() error {

	if s.spec.Success.Kind == "" {
		logrus.Debugf("No shell success criteria defined, updatecli fallbacks to historical workflow")
		s.spec.Success.Kind = "console/output"
	}

	switch s.spec.Success.Kind {
	case "console/output":
		o, err := console.New(&s.result.ExitCode, &s.result.Stdout)
		if err != nil {
			return err
		}

		s.success = o

	case "exitcode":
		o, err := exitcode.New(s.spec.Success.Spec, &s.result.ExitCode, &s.result.Stdout)
		if err != nil {
			return err
		}

		s.success = o

	case "file/checksum":
		o, err := checksum.New(s.spec.Success.Spec, &s.result.ExitCode, &s.result.Stdout)
		if err != nil {
			return err
		}

		s.success = o

	default:
		err := fmt.Errorf("shell success criteria %q is not supported by Updatecli", s.spec.Success.Kind)
		return err
	}

	return nil
}

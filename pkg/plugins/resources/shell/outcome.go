package shell

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/shell/outcome/checksum"
	"github.com/updatecli/updatecli/pkg/plugins/resources/shell/outcome/console"
	"github.com/updatecli/updatecli/pkg/plugins/resources/shell/outcome/exitcode"
)

/*
	Result is the package containing the logic used by the Shell resource resource to identify
	if a shell command outcome should be considered a success, a failure, or a warning
*/

var (
	MappingSpecOutcome = map[string]interface{}{
		"console/output": nil,
		"exitcode":       exitcode.Spec{},
		"file/checksum":  checksum.Spec{},
	}
)

type SpecOutcome struct {
	Kind string                 `yaml:",omitempty"`
	Spec map[string]interface{} `yaml:",omitempty"`
}

type Outcomer interface {
	PreCommand() error
	PostCommand() error
	SourceResult() (string, error)
	ConditionResult() (bool, error)
	TargetResult() (bool, error)
}

func (s *Shell) InitOutcome() error {

	if s.spec.Outcome.Kind == "" {
		logrus.Debugf("No shell outcome defined, updatecli fallbacks to historical workflow")
		s.spec.Outcome.Kind = "console/output"
	}

	switch s.spec.Outcome.Kind {
	case "console/output":
		o, err := console.New(&s.result.ExitCode, &s.result.Stdout)
		if err != nil {
			return err
		}

		s.outcome = o

	case "exitcode":
		o, err := exitcode.New(s.spec.Outcome.Spec, &s.result.ExitCode, &s.result.Stdout)
		if err != nil {
			return err
		}

		s.outcome = o

	case "file/checksum":
		o, err := checksum.New(s.spec.Outcome.Spec, &s.result.ExitCode, &s.result.Stdout)
		if err != nil {
			return err
		}

		s.outcome = o

	default:
		err := fmt.Errorf("shell outcome %q is not supported by Updatecli", s.spec.Outcome.Kind)
		return err
	}

	return nil
}

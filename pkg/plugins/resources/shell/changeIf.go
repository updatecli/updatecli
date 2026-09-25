package shell

import (
	"fmt"

	jschema "github.com/invopop/jsonschema"
	"github.com/updatecli/updatecli/pkg/core/jsonschema"
	"github.com/updatecli/updatecli/pkg/core/result"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/resources/shell/success/checksum"
	"github.com/updatecli/updatecli/pkg/plugins/resources/shell/success/console"
	"github.com/updatecli/updatecli/pkg/plugins/resources/shell/success/exitcode"
)

/*
	Result is the package containing the logic used by the Shell resource to identify
	if a shell command result should be considered a success, a failure, or a warning
*/

const (
	consoleOutputIdentifier = "console/output"
	exitCodeIdentifier      = "exitcode"
	fileChecksumIdentifier  = "file/checksum"
)

var (
	MappingSpecChangedIf = map[string]interface{}{
		consoleOutputIdentifier: &console.Spec{},
		exitCodeIdentifier:      &exitcode.Spec{},
		fileChecksumIdentifier:  &checksum.Spec{},
	}
)

// SpecChangedIf defines how Updatecli interprets the result of a shell command.
type SpecChangedIf struct {
	// "kind" defines the kind of success criteria.
	//
	// default:
	//   console/output
	//
	// remark:
	//   * accepted values are "console/output", "exitcode" and "file/checksum".
	//
	Kind string `yaml:",omitempty"`
	// "spec" defines the parameters of the selected success criteria kind.
	//
	Spec interface{} `yaml:",omitempty"`
}

type Evaluator interface {
	PreCommand(workingDir string) error
	PostCommand(workingDir string) error
	SourceResult(resultSource *result.Source) error
	ConditionResult() (bool, error)
	TargetResult() (bool, error)
}

func (s *Shell) InitChangedIf() error {

	if s.spec.ChangedIf.Kind == "" {
		logrus.Debugf("No shell success criteria defined, updatecli fallbacks to historical workflow")
		s.spec.ChangedIf.Kind = consoleOutputIdentifier
	}

	switch s.spec.ChangedIf.Kind {
	case consoleOutputIdentifier:
		o, err := console.New(&s.result.ExitCode, &s.result.Stdout)
		if err != nil {
			return err
		}

		s.success = o

	case exitCodeIdentifier:
		o, err := exitcode.New(s.spec.ChangedIf.Spec, &s.result.ExitCode, &s.result.Stdout)
		if err != nil {
			return err
		}

		s.success = o

	case fileChecksumIdentifier:
		o, err := checksum.New(s.spec.ChangedIf.Spec, &s.result.ExitCode, &s.result.Stdout)
		if err != nil {
			return err
		}

		s.success = o

	default:
		err := fmt.Errorf("shell success criteria %q is not supported by Updatecli", s.spec.ChangedIf.Kind)
		return err
	}

	return nil
}

// JSONSchema implements the json schema interface to generate the "condition" jsonschema.
func (SpecChangedIf) JSONSchema() *jschema.Schema {
	type SpecSuccessAlias SpecChangedIf
	return jsonschema.AppendOneOfToJsonSchema(SpecSuccessAlias{}, MappingSpecChangedIf)
}

package shell

import (
	"fmt"

	jschema "github.com/invopop/jsonschema"
	"github.com/updatecli/updatecli/pkg/core/jsonschema"

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
		"exitcode":       &exitcode.Spec{},
		"file/checksum":  &checksum.Spec{},
	}
)

type SpecSuccess struct {
	// Kind specifies the success criteria kind, accepted answer ["console/output","exitcode","file/checksum"]
	Kind string `yaml:",omitempty"`
	// Spec specifies the parameter for a specific success criteria kind
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

	if s.spec.ChangedIf.Kind == "" {
		logrus.Debugf("No shell success criteria defined, updatecli fallbacks to historical workflow")
		s.spec.ChangedIf.Kind = "console/output"
	}

	switch s.spec.ChangedIf.Kind {
	case "console/output":
		o, err := console.New(&s.result.ExitCode, &s.result.Stdout)
		if err != nil {
			return err
		}

		s.success = o

	case "exitcode":
		o, err := exitcode.New(s.spec.ChangedIf.Spec, &s.result.ExitCode, &s.result.Stdout)
		if err != nil {
			return err
		}

		s.success = o

	case "file/checksum":
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
func (SpecSuccess) JSONSchema() *jschema.Schema {
	type SpecSuccessAlias SpecSuccess
	return jsonschema.AppendOneOfToJsonSchema(SpecSuccessAlias{}, MappingSpecSuccess)
}

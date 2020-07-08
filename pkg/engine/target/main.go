package target

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/olblak/updateCli/pkg/scm"
	"github.com/olblak/updateCli/pkg/yaml"
)

// Target defines which file needs to be updated based on source output
type Target struct {
	Name    string
	Kind    string
	Prefix  string
	Postfix string
	Spec    interface{}
	Scm     map[string]interface{}
}

// Spec is an interface which offers common function to manipulate targets.
type Spec interface {
	Target() (bool, error)
}

// Check verifies if mandatory Targets parameters are provided and return false if not.
func (t *Target) Check() (bool, error) {
	ok := true
	required := []string{}

	if t.Name == "" {
		required = append(required, "Name")
	}

	if len(required) > 0 {
		err := fmt.Errorf("\u2717 Target parameter(s) required: [%v]", strings.Join(required, ","))
		return false, err
	}

	return ok, nil
}

// Execute applies a specific target configuration
func (t *Target) Execute(source string, o *Options) error {

	var s scm.Scm
	var message string
	var file string

	pwd, err := os.Executable()
	if err != nil {
		panic(err)
	}

	workingDir := filepath.Dir(pwd)

	if len(t.Scm) > 0 {
		_, err := t.Check()
		if err != nil {
			return err
		}

		s, err = scm.Unmarshal(t.Scm)
		if err != nil {
			return err
		}

		err = s.Init(source, t.Name)

		if o.Clean {
			defer s.Clean()
		}

		if err != nil {
			return err
		}

		s.Clone()

		workingDir = s.GetDirectory()

	}

	var spec Spec

	switch t.Kind {

	case "yaml":
		var y yaml.Yaml

		err := mapstructure.Decode(t.Spec, &y)

		if err != nil {
			return err
		}

		y.Value = t.Prefix + source + t.Postfix

		y.Path = workingDir

		file = y.File
		message = fmt.Sprintf("[updatecli] Update %s version to %v\n\nKey '%s', from file '%v', was updated to '%s'\n",
			t.Name,
			y.Value,
			y.Key,
			y.File,
			y.Key)

		spec = &y

	default:
		return fmt.Errorf("Don't support target: %v", t.Kind)
	}

	changed, err := spec.Target()

	if err != nil {
		return err
	}

	if changed {
		if message == "" {
			return fmt.Errorf("Target has no change message")
		}

		if len(t.Scm) > 0 {

			if o.Commit {
				s.Add(file)
				s.Commit(file, message)
			}
			if o.Push {
				s.Push()
			}
		}
	}

	return nil
}

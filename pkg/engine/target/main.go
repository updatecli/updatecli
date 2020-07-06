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
	GetFile() string
	Target(source, name string, workDir string) (bool, string, error)
}

// Unmarshal parses target spec and return its Spec interface
func (t *Target) Unmarshal() (Spec, error) {

	var spec Spec

	switch t.Kind {

	case "yaml":
		var y yaml.Yaml

		err := mapstructure.Decode(t.Spec, &y)

		if err != nil {
			return nil, err
		}

		spec = &y

	default:
		return nil, fmt.Errorf("⚠ Don't support target: %v", t.Kind)
	}

	return spec, nil
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

	_, err := t.Check()
	if err != nil {
		return err
	}

	scm, err := scm.Unmarshal(t.Scm)
	if err != nil {
		return err
	}

	spec, err := t.Unmarshal()

	if err != nil {
		return err
	}

	err = scm.Init(source, t.Name)

	if o.Clean {
		defer scm.Clean()
	}

	if err != nil {
		return err
	}

	file := spec.GetFile()

	path := filepath.Join(scm.GetDirectory(), file)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		scm.Clone()
	}

	changed, message, err := spec.Target(
		t.Prefix+source+t.Postfix,
		t.Name,
		scm.GetDirectory())

	if err != nil {
		return err
	}

	if changed {
		if message == "" {
			return fmt.Errorf("Target has no change message")
		}

		if o.Commit {
			scm.Add(file)
			scm.Commit(file, message)
		}
		if o.Push {
			scm.Push()
		}
	}

	return nil
}

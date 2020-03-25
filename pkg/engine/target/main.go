package target

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/mapstructure"
	"github.com/olblak/updateCli/pkg/scm"
	"github.com/olblak/updateCli/pkg/yaml"
)

// Target defines which file needs to be updated based on source output
type Target struct {
	Name string
	Kind string
	Spec interface{}
	Scm  map[string]interface{}
}

// Spec is an interface which offers common function to manipulate targets.
type Spec interface {
	GetFile() string
	Target(source, workDir string) (bool, string, error)
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

// Execute applies a specific target configuration
func (t *Target) Execute(source string) error {

	scm, err := scm.Unmarshal(t.Scm)
	if err != nil {
		return err
	}

	spec, err := t.Unmarshal()

	if err != nil {
		return err
	}

	err = scm.Init(source)
	if err != nil {
		return err
	}

	file := spec.GetFile()

	path := filepath.Join(scm.GetDirectory(), file)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		scm.Clone()
	}

	changed, message, err := spec.Target(source, scm.GetDirectory())

	if err != nil {
		return err
	}

	if changed {
		if message == "" {
			return fmt.Errorf("Target has no change message")
		}
		scm.Add(file)
		scm.Commit(file, message)
		scm.Push()
		scm.Clean()
	}

	return nil
}

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

	pwd, err := os.Getwd()
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

		if dir, base, err := isFileExist(y.File); y.Path == "" && err == nil {
			y.Path = dir
			y.File = base
		} else if !isDirectory(y.Path) {
			fmt.Printf("Directory %s is not valid so fallback to current directory %s\n", y.Path, workingDir)
			y.Path = workingDir
		} else {
			fmt.Println("Something weird happened while settings yaml directory")
		}

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

func isFileExist(file string) (dir string, base string, err error) {
	if _, err := os.Stat(file); err != nil {
		return "", "", err
	}

	absolutePath, err := filepath.Abs(file)
	if err != nil {
		return "", "", err
	}
	dir = filepath.Dir(absolutePath)
	base = filepath.Base(absolutePath)

	return dir, base, err
}

func isDirectory(path string) bool {

	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return true
	}
	return false
}

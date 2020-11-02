package target

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/olblak/updateCli/pkg/core/scm"
	"github.com/olblak/updateCli/pkg/plugins/yaml"
)

// Target defines which file needs to be updated based on source output
type Target struct {
	Name      string
	Kind      string
	Changelog string `yaml:"-"`
	Prefix    string
	Postfix   string
	Spec      interface{}
	Scm       map[string]interface{}
	Result    string `yaml:"-"`
}

// Spec is an interface which offers common function to manipulate targets.
type Spec interface {
	Target(source string, dryRun bool) (bool, error)
	TargetFromSCM(source string, scm scm.Scm, dryRun bool) (changed bool, files []string, message string, err error)
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

// Unmarshal decodes a target struct
func Unmarshal(target *Target) (spec Spec, err error) {
	switch target.Kind {
	case "yaml":
		y := yaml.Yaml{}

		err := mapstructure.Decode(target.Spec, &y)

		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		spec = &y
	}
	return spec, nil
}

// Run applies a specific target configuration
func (t *Target) Run(source string, o *Options) (changed bool, err error) {

	if o.DryRun {

		fmt.Printf("**Dry Run enabled**\n\n")
	}

	spec, err := Unmarshal(t)

	if err != nil {
		return false, err
	}

	if err != nil {
		return false, err
	}

	if len(t.Scm) > 0 {
		var message string
		var files []string
		var s scm.Scm

		_, err := t.Check()
		if err != nil {
			return false, err
		}

		s, err = scm.Unmarshal(t.Scm)
		if err != nil {
			return false, err
		}

		err = s.Init(source, t.Name)

		if err != nil {
			return false, err
		}

		changed, files, message, err = spec.TargetFromSCM(t.Prefix+source+t.Postfix, s, o.DryRun)

		if err != nil {
			return changed, err
		}

		if changed && !o.DryRun {
			if message == "" {
				return changed, fmt.Errorf("Target has no change message")
			}

			if len(t.Scm) > 0 {

				if o.Commit {
					s.Add(files)
					s.Commit(message)
				}
				if o.Push {
					s.Push()
				}
			}
		}

	} else if len(t.Scm) == 0 {

		changed, err = spec.Target(t.Prefix+source+t.Postfix, o.DryRun)

		if err != nil {
			return changed, err
		}

	}

	return changed, nil
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

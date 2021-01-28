package target

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/olblak/updateCli/pkg/core/scm"
	"github.com/olblak/updateCli/pkg/plugins/docker/dockerfile"
	"github.com/olblak/updateCli/pkg/plugins/file"
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
	case "dockerfile":
		d := dockerfile.Dockerfile{}

		err := mapstructure.Decode(target.Spec, &d)
		if err != nil {
			logrus.Errorf("err - %s", err)
			return nil, err
		}

		spec = &d

	case "yaml":
		y := yaml.Yaml{}

		err := mapstructure.Decode(target.Spec, &y)

		if err != nil {
			logrus.Errorf("err - %s", err)
			return nil, err
		}

		spec = &y

	case "file":
		f := file.File{}

		err := mapstructure.Decode(target.Spec, &f)

		if err != nil {
			logrus.Errorf("err - %s", err)
			return nil, err
		}

		spec = &f

	default:
		return nil, fmt.Errorf("⚠ Don't support target kind: %v", target.Kind)
	}
	return spec, nil
}

// Run applies a specific target configuration
func (t *Target) Run(source string, o *Options) (changed bool, err error) {

	if o.DryRun {

		logrus.Infof("**Dry Run enabled**\n")
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

		err = s.Checkout()
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
					err := s.Add(files)
					if err != nil {
						return changed, err
					}

					err = s.Commit(message)
					if err != nil {
						return changed, err
					}
				}
				if o.Push {
					err := s.Push()
					if err != nil {
						return changed, err
					}
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

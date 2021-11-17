package target

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/scm"
	"github.com/updatecli/updatecli/pkg/core/transformer"
	"github.com/updatecli/updatecli/pkg/plugins/docker/dockerfile"
	"github.com/updatecli/updatecli/pkg/plugins/file"
	"github.com/updatecli/updatecli/pkg/plugins/git/tag"
	"github.com/updatecli/updatecli/pkg/plugins/helm/chart"
	"github.com/updatecli/updatecli/pkg/plugins/shell"
	"github.com/updatecli/updatecli/pkg/plugins/yaml"
)

// Target defines which file needs to be updated based on source output
type Target struct {
	Result      string // Result store the condition result after a target run. This variable can't be set by an updatecli configuration
	Config      Config
	ReportBody  string
	ReportTitle string
	Changelog   string
	Commit      bool
	Push        bool
	Clean       bool
	DryRun      bool
}

// Config defines target parameters
type Config struct {
	DependsOn    []string `yaml:"depends_on"`
	Name         string
	PipelineID   string `yaml:"pipelineID"` // PipelineID references a uniq pipeline run that allows to groups targets
	Kind         string
	Prefix       string // Deprecated in favor of Transformers on 2021/01/3
	Postfix      string // Deprecated in favor of Transformers on 2021/01/3
	ReportTitle  string // ReportTitle contains the updatecli reports title for sources and conditions run
	ReportBody   string // ReportBody contains the updatecli reports body for sources and conditions run
	Transformers transformer.Transformers
	Spec         interface{}
	Scm          map[string]interface{}
	SourceID     string `yaml:"sourceID"`
}

// Targeter is an interface which offers common function to manipulate targets.
type Targeter interface {
	Target(source string, dryRun bool) (bool, error)
	TargetFromSCM(source string, scm scm.Scm, dryRun bool) (changed bool, files []string, message string, err error)
}

// Check verifies if mandatory Targets parameters are provided and return false if not.
func (t *Target) Check() (bool, error) {
	ok := true
	required := []string{}

	if t.Config.Name == "" {
		required = append(required, "Name")
	}

	if len(required) > 0 {
		err := fmt.Errorf("\u2717 Target parameter(s) required: [%v]", strings.Join(required, ","))
		return false, err
	}

	return ok, nil
}

// Unmarshal decodes a target struct
func Unmarshal(target *Target) (targeter Targeter, err error) {
	switch target.Config.Kind {
	case "helmChart":
		ch := chart.Chart{}

		err := mapstructure.Decode(target.Config.Spec, &ch)

		if err != nil {
			logrus.Errorf("err - %s", err)
			return nil, err
		}

		targeter = &ch

	case "dockerfile":
		d := dockerfile.Dockerfile{}

		err := mapstructure.Decode(target.Config.Spec, &d)
		if err != nil {
			logrus.Errorf("err - %s", err)
			return nil, err
		}

		targeter = &d

	case "gitTag":
		t := tag.Tag{}

		err := mapstructure.Decode(target.Config.Spec, &t)
		if err != nil {
			logrus.Errorf("err - %s", err)
			return nil, err
		}

		targeter = &t

	case "yaml":
		var targetSpec yaml.YamlSpec

		if err := mapstructure.Decode(target.Config.Spec, &targetSpec); err != nil {
			return nil, err
		}

		targeter, err = yaml.New(targetSpec)
		if err != nil {
			return nil, err
		}

	case "file":
		var targetSpec file.FileSpec

		if err := mapstructure.Decode(target.Config.Spec, &targetSpec); err != nil {
			return nil, err
		}

		targeter, err = file.New(targetSpec)
		if err != nil {
			return nil, err
		}

	case "shell":
		shellResourceSpec := shell.ShellSpec{}

		err := mapstructure.Decode(target.Config.Spec, &shellResourceSpec)
		if err != nil {
			return nil, err
		}

		targeter, err = shell.New(shellResourceSpec)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("⚠ Don't support target kind: %v", target.Config.Kind)
	}
	return targeter, nil
}

// Run applies a specific target configuration
func (t *Target) Run(source string, o *Options) (err error) {

	var pr scm.PullRequest
	var changed bool

	if len(t.Config.Transformers) > 0 {
		source, err = t.Config.Transformers.Apply(source)
		if err != nil {
			t.Result = result.FAILURE
			logrus.Error(err)
			return err
		}
	}

	// Announce deprecation on 2021/01/31
	if len(t.Config.Prefix) > 0 {
		logrus.Warnf("Key 'prefix' deprecated in favor of 'transformers', it will be delete in a future release")
	}

	// Announce deprecation on 2021/01/31
	if len(t.Config.Postfix) > 0 {
		logrus.Warnf("Key 'postfix' deprecated in favor of 'transformers', it will be delete in a future release")
	}

	if o.DryRun {
		logrus.Infof("\n**Dry Run enabled**\n\n")
	}

	spec, err := Unmarshal(t)

	if err != nil {
		t.Result = result.FAILURE
		return err
	}

	if err != nil {
		t.Result = result.FAILURE
		return err
	}

	if len(t.Config.Scm) == 0 {

		changed, err = spec.Target(t.Config.Prefix+source+t.Config.Postfix, o.DryRun)
		if err != nil {
			t.Result = result.FAILURE
			return err
		}

		if changed {
			t.Result = result.ATTENTION
		} else {
			t.Result = result.SUCCESS
		}
		return nil

	}

	var message string
	var files []string
	var s scm.Scm

	_, err = t.Check()
	if err != nil {
		t.Result = result.FAILURE
		return err
	}

	s, pr, err = scm.Unmarshal(t.Config.Scm)
	if err != nil {
		t.Result = result.FAILURE
		return err
	}

	if err = s.Init(source, t.Config.PipelineID); err != nil {
		t.Result = result.FAILURE
		return err
	}

	if err = s.Checkout(); err != nil {
		t.Result = result.FAILURE
		return err
	}

	changed, files, message, err = spec.TargetFromSCM(t.Config.Prefix+source+t.Config.Postfix, s, o.DryRun)
	if err != nil {
		t.Result = result.FAILURE
		return err
	}

	if !changed {
		t.Result = result.SUCCESS
		return nil
	}

	t.Result = result.ATTENTION
	if !o.DryRun {
		if message == "" {
			t.Result = result.FAILURE
			return fmt.Errorf("target has no change message")
		}

		if len(t.Config.Scm) > 0 {

			if len(files) == 0 {
				t.Result = result.FAILURE
				logrus.Info("no changed files to commit")
				return nil
			}

			if o.Commit {
				if err := s.Add(files); err != nil {
					t.Result = result.FAILURE
					return err
				}

				if err = s.Commit(message); err != nil {
					t.Result = result.FAILURE
					return err
				}
			}
			if o.Push {
				if err := s.Push(); err != nil {
					t.Result = result.FAILURE
					return err
				}
				if pr != nil {
					err = pr.CreatePullRequest(
						t.ReportTitle,
						t.Changelog,
						t.ReportBody,
					)
					if err != nil {
						t.Result = result.FAILURE
						return err
					}
				}
			}
		}
	}

	return nil
}

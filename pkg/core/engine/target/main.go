package target

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/mitchellh/mapstructure"
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
	DependsOn    []string `yaml:"depends_on"`
	Name         string
	PipelineID   string `yaml:"pipelineID"` // PipelineID references a uniq pipeline run that allows to groups targets
	Kind         string
	Changelog    string
	Prefix       string // Deprecated in favor of Transformers on 2021/01/3
	Postfix      string // Deprecated in favor of Transformers on 2021/01/3
	ReportTitle  string // ReportTitle contains the updatecli reports title for sources and conditions run
	ReportBody   string // ReportBody contains the updatecli reports body for sources and conditions run
	Transformers transformer.Transformers
	Spec         interface{}
	Scm          map[string]interface{}
	SourceID     string `yaml:"sourceID"`
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
	case "helmChart":
		ch := chart.Chart{}

		err := mapstructure.Decode(target.Spec, &ch)

		if err != nil {
			logrus.Errorf("err - %s", err)
			return nil, err
		}

		spec = &ch

	case "dockerfile":
		d := dockerfile.Dockerfile{}

		err := mapstructure.Decode(target.Spec, &d)
		if err != nil {
			logrus.Errorf("err - %s", err)
			return nil, err
		}

		spec = &d

	case "gitTag":
		t := tag.Tag{}

		err := mapstructure.Decode(target.Spec, &t)
		if err != nil {
			logrus.Errorf("err - %s", err)
			return nil, err
		}

		spec = &t

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

	case "shell":
		shellResourceSpec := shell.ShellSpec{}

		err := mapstructure.Decode(target.Spec, &shellResourceSpec)
		if err != nil {
			return nil, err
		}

		spec, err = shell.New(shellResourceSpec)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("⚠ Don't support target kind: %v", target.Kind)
	}
	return spec, nil
}

// Run applies a specific target configuration
func (t *Target) Run(source string, o *Options) (changed bool, err error) {

	var pr scm.PullRequest

	if len(t.Transformers) > 0 {
		source, err = t.Transformers.Apply(source)
		if err != nil {
			logrus.Error(err)
			return false, err
		}
	}

	// Announce deprecation on 2021/01/31
	if len(t.Prefix) > 0 {
		logrus.Warnf("Key 'prefix' deprecated in favor of 'transformers', it will be delete in a future release")
	}

	// Announce deprecation on 2021/01/31
	if len(t.Postfix) > 0 {
		logrus.Warnf("Key 'postfix' deprecated in favor of 'transformers', it will be delete in a future release")
	}

	if o.DryRun {

		logrus.Infof("\n**Dry Run enabled**\n\n")
	}

	spec, err := Unmarshal(t)

	if err != nil {
		return false, err
	}

	if err != nil {
		return false, err
	}

	if len(t.Scm) == 0 {

		changed, err = spec.Target(t.Prefix+source+t.Postfix, o.DryRun)
		if err != nil {
			return changed, err
		}
		return changed, nil

	}

	var message string
	var files []string
	var s scm.Scm

	_, err = t.Check()
	if err != nil {
		return false, err
	}

	s, pr, err = scm.Unmarshal(t.Scm)
	if err != nil {
		return false, err
	}

	err = s.Init(source, t.PipelineID)
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

			if len(files) == 0 {
				logrus.Info("no changed files to commit")
				return changed, nil
			}

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
	if pr != nil && !o.DryRun && o.Push {
		pr.InitPullRequestDescription(
			t.ReportTitle,
			t.Changelog,
			t.ReportBody,
		)

		ID, err := pr.IsPullRequest()

		logrus.Debugf("Pull Request ID: %v", ID)

		// We try to update the pullrequest even if nothing changed in the current run
		// so we always update the pull request body if needed.
		if err != nil {
			return changed, err
		} else if len(ID) == 0 && err == nil && changed {
			// Something changed and no open pull request exist so we create it
			logrus.Infof("Creating Pull Request\n")
			err = pr.OpenPullRequest()
			if err != nil {
				return changed, err
			}

		} else if len(ID) != 0 && err == nil {
			// A pull request already exist so we try to update it to overide old information
			// whatever a change was applied during the run or not
			logrus.Infof("Pull Request already exist, updating it\n")
			err = pr.UpdatePullRequest(ID)
			if err != nil {
				return changed, err
			}
			// No change and no pull request, nothing else to do
		} else if len(ID) == 0 && err == nil && !changed {
			// We try to catch un-planned scenario
		} else {
			logrus.Errorf("Something unexpected happened while dealing with Pull Request")
		}
	}

	return changed, nil
}

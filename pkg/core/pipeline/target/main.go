package target

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/dockerfile"
	"github.com/updatecli/updatecli/pkg/plugins/file"
	"github.com/updatecli/updatecli/pkg/plugins/gittag"
	"github.com/updatecli/updatecli/pkg/plugins/helm"
	"github.com/updatecli/updatecli/pkg/plugins/shell"
	"github.com/updatecli/updatecli/pkg/plugins/yaml"
)

// Target defines which file needs to be updated based on source output
type Target struct {
	Result string // Result store the condition result after a target run. This variable can't be set by an updatecli configuration
	Config Config
	Commit bool
	Push   bool
	Clean  bool
	DryRun bool
	Scm    *scm.ScmHandler
}

// Config defines target parameters
type Config struct {
	resource.ResourceConfig `yaml:",inline"`
	PipelineID              string `yaml:"pipelineID"` // PipelineID references a unique pipeline run allowing to group targets
	ReportTitle             string // ReportTitle contains the updatecli reports title for sources and conditions run
	ReportBody              string // ReportBody contains the updatecli reports body for sources and conditions run
	SourceID                string `yaml:"sourceID"`
}

// Targeter is an interface which offers common function to manipulate targets.
type Targeter interface {
	Target(source string, dryRun bool) (bool, error)
	TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (changed bool, files []string, message string, err error)
}

// Check verifies if mandatory Targets parameters are provided and return false if not.
func (t *Target) Check() (bool, error) {
	ok := true
	required := []string{}

	if t.Config.Name == "" {
		required = append(required, "Name")
	}

	if len(required) > 0 {
		err := fmt.Errorf("%s Target parameter(s) required: [%v]", result.FAILURE, strings.Join(required, ","))
		return false, err
	}

	return ok, nil
}

// Unmarshal decodes a target struct
func Unmarshal(target *Target) (targeter Targeter, err error) {
	switch target.Config.Kind {
	case "helmChart":
		targetSpec := helm.Spec{}

		err := mapstructure.Decode(target.Config.Spec, &targetSpec)
		if err != nil {
			return nil, err
		}

		targeter, err = helm.New(targetSpec)
		if err != nil {
			return nil, err
		}

	case "dockerfile":
		targetSpec := dockerfile.Spec{}

		err := mapstructure.Decode(target.Config.Spec, &targetSpec)
		if err != nil {
			return nil, err
		}

		targeter, err = dockerfile.New(targetSpec)
		if err != nil {
			return nil, err
		}

	case "gitTag":
		targetSpec := gittag.Spec{}

		err := mapstructure.Decode(target.Config.Spec, &targetSpec)
		if err != nil {
			return nil, err
		}

		targeter, err = gittag.New(targetSpec)
		if err != nil {
			return nil, err
		}

	case "yaml":
		var targetSpec yaml.Spec

		if err := mapstructure.Decode(target.Config.Spec, &targetSpec); err != nil {
			return nil, err
		}

		targeter, err = yaml.New(targetSpec)
		if err != nil {
			return nil, err
		}

	case "file":
		var targetSpec file.Spec

		if err := mapstructure.Decode(target.Config.Spec, &targetSpec); err != nil {
			return nil, err
		}

		targeter, err = file.New(targetSpec)
		if err != nil {
			return nil, err
		}

	case "shell":
		targetSpec := shell.Spec{}

		err := mapstructure.Decode(target.Config.Spec, &targetSpec)
		if err != nil {
			return nil, err
		}

		targeter, err = shell.New(targetSpec)
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

	var changed bool

	if len(t.Config.Transformers) > 0 {
		source, err = t.Config.Transformers.Apply(source)
		if err != nil {
			t.Result = result.FAILURE
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

	// If no scm configuration provided then stop early
	if t.Scm == nil {

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

	_, err = t.Check()
	if err != nil {
		t.Result = result.FAILURE
		return err
	}

	s := *t.Scm

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

		if len(files) == 0 {
			t.Result = result.FAILURE
			logrus.Info("no changed file to commit")
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
		}
	}

	return nil
}

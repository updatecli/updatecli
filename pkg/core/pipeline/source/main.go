package source

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/awsami"
	"github.com/updatecli/updatecli/pkg/plugins/dockerdigest"
	"github.com/updatecli/updatecli/pkg/plugins/file"
	"github.com/updatecli/updatecli/pkg/plugins/github"
	"github.com/updatecli/updatecli/pkg/plugins/githubrelease"
	"github.com/updatecli/updatecli/pkg/plugins/gittag"
	"github.com/updatecli/updatecli/pkg/plugins/helm"
	"github.com/updatecli/updatecli/pkg/plugins/jenkins"
	"github.com/updatecli/updatecli/pkg/plugins/maven"
	"github.com/updatecli/updatecli/pkg/plugins/shell"
	"github.com/updatecli/updatecli/pkg/plugins/version"
	"github.com/updatecli/updatecli/pkg/plugins/yaml"
)

// Source defines how a value is retrieved from a specific source
type Source struct {
	Changelog string // Changelog hold the changelog description
	Result    string // Result store the source result after a source run. This variable can't be set by an updatecli configuration
	Output    string // Output contains the value retrieved from a source
	Config    Config // Config defines a source specifications
	Scm       *scm.ScmHandler
}

// Config struct defines a source configuration
type Config struct {
	resource.ResourceConfig `yaml:",inline"`
	// Deprecated in favor of Transformers on 2021/01/3
	Replaces Replacers
}

// Sourcer source is an interface to handle source spec
type Sourcer interface {
	Source(workingDir string) (string, error)
}

// Run execute actions defined by the source configuration
func (s *Source) Run() (err error) {

	spec, changelog, err := s.Unmarshal()

	if err != nil {
		s.Result = result.FAILURE
		return err
	}

	workingDir := ""

	if s.Scm != nil {

		SCM := *s.Scm

		if err != nil {
			s.Result = result.FAILURE
			return err
		}

		err = SCM.Init("", workingDir)

		if err != nil {
			s.Result = result.FAILURE
			return err
		}

		err = SCM.Checkout()

		if err != nil {
			s.Result = result.FAILURE
			return err
		}

		workingDir = SCM.GetDirectory()

	} else if s.Scm == nil {

		pwd, err := os.Getwd()
		if err != nil {
			s.Result = result.FAILURE
			return err
		}

		workingDir = pwd
	}

	s.Output, err = spec.Source(workingDir)
	s.Result = result.SUCCESS

	if err != nil {
		s.Result = result.FAILURE
		return err
	}

	// Retrieve changelog using default source output before
	// modifying its value with the transformer
	if changelog != nil {
		s.Changelog, err = changelog.Changelog(version.Version{
			OriginalVersion: s.Output,
			ParsedVersion:   s.Output,
		})
		if err != nil {
			s.Result = result.FAILURE
			// Changelog information are not important enough to fail a pipeline
			logrus.Errorln(err)
		}
	}

	if len(s.Config.Transformers) > 0 {
		s.Output, err = s.Config.Transformers.Apply(s.Output)
		if err != nil {
			s.Result = result.FAILURE
			return err
		}
	}

	// Announce deprecation on 2021/01/31
	if len(s.Config.Prefix) > 0 {
		logrus.Warnf("Key 'prefix' deprecated in favor of 'transformers', it will be delete in a future release\n")
	}

	// Announce deprecation on 2021/01/31
	if len(s.Config.Postfix) > 0 {
		logrus.Warnf("Key 'postfix' deprecated in favor of 'transformers', it will be delete in a future release\n")
	}

	// Deprecated in favor of Transformers on 2021/01/3
	if len(s.Config.Replaces) > 0 {
		args := s.Config.Replaces.Unmarshal()

		r := strings.NewReplacer(args...)
		s.Output = (r.Replace(s.Output))
	}

	if len(s.Output) == 0 {
		s.Result = result.ATTENTION
	}

	return err
}

// Unmarshal decode a source spec and returned its typed content
func (s *Source) Unmarshal() (sourcer Sourcer, changelog Changelog, err error) {
	switch s.Config.Kind {
	case "aws/ami":
		var sourceSpec awsami.Spec

		if err := mapstructure.Decode(s.Config.Spec, &sourceSpec); err != nil {
			return nil, nil, err
		}

		sourcer, err = awsami.New(sourceSpec)
		if err != nil {
			return nil, nil, err
		}

	case "dockerDigest":
		var sourceSpec dockerdigest.Spec

		if err := mapstructure.Decode(s.Config.Spec, &sourceSpec); err != nil {
			return nil, nil, err
		}

		sourcer, err = dockerdigest.New(sourceSpec)
		if err != nil {
			return nil, nil, err
		}

	case "githubRelease":
		var sourceSpec github.Spec

		if err := mapstructure.Decode(s.Config.Spec, &sourceSpec); err != nil {
			return nil, nil, err
		}

		newGhrResource, err := githubrelease.New(sourceSpec)
		if err != nil {
			return nil, nil, err
		}
		sourcer = &newGhrResource
		changelog = &newGhrResource

	case "file":
		var sourceSpec file.Spec

		if err := mapstructure.Decode(s.Config.Spec, &sourceSpec); err != nil {
			return nil, nil, err
		}

		sourcer, err = file.New(sourceSpec)
		if err != nil {
			return nil, nil, err
		}

	case "helmChart":
		var sourceSpec helm.Spec

		if err := mapstructure.Decode(s.Config.Spec, &sourceSpec); err != nil {
			return nil, nil, err
		}

		helmChartResource, err := helm.New(sourceSpec)
		if err != nil {
			return nil, nil, err
		}
		sourcer = helmChartResource
		changelog = helmChartResource

	case "jenkins":
		var sourceSpec jenkins.Spec

		if err := mapstructure.Decode(s.Config.Spec, &sourceSpec); err != nil {
			return nil, nil, err
		}

		jenkinsResource, err := jenkins.New(sourceSpec)
		if err != nil {
			return nil, nil, err
		}
		sourcer = &jenkinsResource
		changelog = &jenkinsResource

	case "maven":
		var sourceSpec maven.Spec

		if err := mapstructure.Decode(s.Config.Spec, &sourceSpec); err != nil {
			return nil, nil, err
		}

		sourcer, err = maven.New(sourceSpec)
		if err != nil {
			return nil, nil, err
		}

	case "gitTag":
		var sourceSpec gittag.Spec

		if err := mapstructure.Decode(s.Config.Spec, &sourceSpec); err != nil {
			return nil, nil, err
		}

		sourcer, err = gittag.New(sourceSpec)
		if err != nil {
			return nil, nil, err
		}

	case "yaml":
		var sourceSpec yaml.Spec

		if err := mapstructure.Decode(s.Config.Spec, &sourceSpec); err != nil {
			return nil, nil, err
		}

		sourcer, err = yaml.New(sourceSpec)
		if err != nil {
			return nil, nil, err
		}

	case "shell":
		shellResourceSpec := shell.Spec{}

		err := mapstructure.Decode(s.Config.Spec, &shellResourceSpec)
		if err != nil {
			return nil, nil, err
		}

		sourcer, err = shell.New(shellResourceSpec)
		if err != nil {
			return nil, nil, err
		}

	default:
		return nil, nil, fmt.Errorf("⚠ Don't support source kind: %v", s.Config.Kind)
	}
	return sourcer, changelog, nil

}

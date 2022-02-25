package resource

import (
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/transformer"
	"github.com/updatecli/updatecli/pkg/plugins/resources/awsami"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerdigest"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerfile"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
	"github.com/updatecli/updatecli/pkg/plugins/resources/file"
	"github.com/updatecli/updatecli/pkg/plugins/resources/githubrelease"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gittag"
	"github.com/updatecli/updatecli/pkg/plugins/resources/helm"
	"github.com/updatecli/updatecli/pkg/plugins/resources/jenkins"
	"github.com/updatecli/updatecli/pkg/plugins/resources/maven"
	"github.com/updatecli/updatecli/pkg/plugins/resources/shell"
	"github.com/updatecli/updatecli/pkg/plugins/resources/yaml"
)

type ResourceConfig struct {
	DependsOn []string `yaml:"depends_on"`
	Name      string
	Kind      string
	// Deprecated in favor of Transformers on 2021/01/3
	Prefix string
	// Deprecated in favor of Transformers on 2021/01/3
	Postfix      string
	Transformers transformer.Transformers
	Spec         interface{}
	// Deprecated field on version [1.17.0]
	Scm   map[string]interface{}
	SCMID string `yaml:"scmID"` // SCMID references a uniq scm configuration
}

// Implements the Unmarshaler interface of the yaml pkg.
func (r *ResourceConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	yamlResourceConfig := ResourceConfig{}
	err := unmarshal(&yamlResourceConfig)
	if err != nil {
		return err
	}

	switch strings.ToLower(yamlResourceConfig.Kind) {
	case "aws/ami":
		a, err := awsami.New(yamlResourceConfig.Spec)

		if err != nil {
			return err
		}

		yamlResourceConfig.Spec = a.Spec

	//case "dockerdigest":
	//	return dockerdigest.New(rs.Spec)
	//case "dockerfile":
	//	return dockerfile.New(rs.Spec)
	//case "dockerimage":
	//	return dockerimage.New(rs.Spec)
	//case "githubrelease":
	//	return githubrelease.New(rs.Spec)
	//case "gittag":
	//	return gittag.New(rs.Spec)
	//case "file":
	//	return file.New(rs.Spec)
	//case "helmchart":
	//	return helm.New(rs.Spec)
	//case "jenkins":
	//	return jenkins.New(rs.Spec)
	//case "maven":
	//	return maven.New(rs.Spec)
	//case "shell":
	//	return shell.New(rs.Spec)
	//case "yaml":
	//	return yaml.New(rs.Spec)
	default:
		return fmt.Errorf("⚠ don't support source kind: %v", yamlResourceConfig.Kind)
	}

	//// make sure to dereference before assignment,
	//// otherwise only the local variable will be overwritten
	//// and not the value the pointer actually points to
	*r = yamlResourceConfig

	return nil
}

// New returns a newly initialized Resource or an error
func New(rs ResourceConfig) (resource Resource, err error) {
	switch strings.ToLower(rs.Kind) {
	case "aws/ami":
		return awsami.New(rs.Spec)
	case "dockerdigest":
		return dockerdigest.New(rs.Spec)
	case "dockerfile":
		return dockerfile.New(rs.Spec)
	case "dockerimage":
		return dockerimage.New(rs.Spec)
	case "githubrelease":
		return githubrelease.New(rs.Spec)
	case "gittag":
		return gittag.New(rs.Spec)
	case "file":
		return file.New(rs.Spec)
	case "helmchart":
		return helm.New(rs.Spec)
	case "jenkins":
		return jenkins.New(rs.Spec)
	case "maven":
		return maven.New(rs.Spec)
	case "shell":
		return shell.New(rs.Spec)
	case "yaml":
		return yaml.New(rs.Spec)
	default:
		return nil, fmt.Errorf("⚠ Don't support resource kind: %v", rs.Kind)
	}
}

// Resource allow to manipulate a resource that can be a source, a condition or a target
type Resource interface {
	Source(workingDir string) (string, error)
	Condition(version string) (bool, error)
	ConditionFromSCM(version string, scm scm.ScmHandler) (bool, error)
	Target(source string, dryRun bool) (bool, error)
	TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (changed bool, files []string, message string, err error)
	Changelog() string
}

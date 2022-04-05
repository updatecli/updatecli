package resource

import (
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
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
	// depends_on specifies which resources must be executed before the current one
	DependsOn []string `yaml:"depends_on"`
	// name specifies the resource name
	Name string
	// kind specifies the resource kind which defines accepted spec value
	Kind string
	// transformers defines how the default input value need to be transformed
	Transformers transformer.Transformers
	// spec specifies parameters for a specific resource kind
	Spec interface{}
	// Deprecated field on version [1.17.0]
	Scm map[string]interface{}
	// scmid specifies the scm configuration key associated to the current resource
	SCMID string // SCMID references a uniq scm configuration
	// !deprecated, please use scmid
	DeprecatedSCMID string `yaml:"scmID"` // SCMID references a uniq scm configuration
}

// New returns a newly initialized Resource or an error
func New(rs ResourceConfig) (resource Resource, err error) {

	kind := strings.ToLower(rs.Kind)

	if _, ok := GetResourceMapping()[kind]; !ok {
		return nil, fmt.Errorf("⚠ Don't support resource kind: %v", rs.Kind)
	}

	switch kind {
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
		return nil, fmt.Errorf("%s Don't support resource kind: %v", result.FAILURE, rs.Kind)
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

// Need to do reflect of ResourceConfig
func GetResourceMapping() map[string]interface{} {

	return map[string]interface{}{
		"aws/ami":       &awsami.Spec{},
		"jenkins":       &jenkins.Spec{},
		"shell":         &shell.Spec{},
		"gittag":        &gittag.Spec{},
		"githubrelease": &githubrelease.Spec{},
		"dockerdigest":  &dockerdigest.Spec{},
		"dockerfile":    &dockerfile.Spec{},
		"dockerimage":   &dockerimage.Spec{},
		"file":          &file.Spec{},
		"helmchart":     &helm.Spec{},
		"maven":         &maven.Spec{},
		"yaml":          &yaml.Spec{},
	}
}

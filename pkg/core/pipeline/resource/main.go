package resource

import (
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/transformer"
	"github.com/updatecli/updatecli/pkg/plugins/resources/awsami"
	"github.com/updatecli/updatecli/pkg/plugins/resources/csv"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerdigest"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerfile"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerimage"
	"github.com/updatecli/updatecli/pkg/plugins/resources/file"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitbranch"
	giteaBranch "github.com/updatecli/updatecli/pkg/plugins/resources/gitea/branch"
	giteaRelease "github.com/updatecli/updatecli/pkg/plugins/resources/gitea/release"
	giteaTag "github.com/updatecli/updatecli/pkg/plugins/resources/gitea/tag"
	"github.com/updatecli/updatecli/pkg/plugins/resources/githubrelease"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gittag"
	"github.com/updatecli/updatecli/pkg/plugins/resources/helm"
	"github.com/updatecli/updatecli/pkg/plugins/resources/jenkins"
	"github.com/updatecli/updatecli/pkg/plugins/resources/json"
	"github.com/updatecli/updatecli/pkg/plugins/resources/maven"
	"github.com/updatecli/updatecli/pkg/plugins/resources/npm"
	"github.com/updatecli/updatecli/pkg/plugins/resources/shell"
	"github.com/updatecli/updatecli/pkg/plugins/resources/toml"
	"github.com/updatecli/updatecli/pkg/plugins/resources/xml"
	"github.com/updatecli/updatecli/pkg/plugins/resources/yaml"
)

type ResourceConfig struct {
	// dependson specifies which resources must be executed before the current one
	DependsOn []string `yaml:",omitempty"`
	// name specifies the resource name
	Name string `yaml:",omitempty"`
	// kind specifies the resource kind which defines accepted spec value
	Kind string `yaml:",omitempty" jsonschema:"required"`
	// transformers defines how the default input value need to be transformed
	Transformers transformer.Transformers `yaml:",omitempty"`
	// spec specifies parameters for a specific resource kind
	Spec interface{} `yaml:",omitempty"`
	// scmid specifies the scm configuration key associated to the current resource
	SCMID string `yaml:",omitempty"` // SCMID references a uniq scm configuration
	// !deprecated, please use scmid
	DeprecatedSCMID string `yaml:"scmID,omitempty" jsonschema:"-"` // SCMID references a uniq scm configuration
	// !deprecated, please use dependson
	DeprecatedDependsOn []string `yaml:"depends_on,omitempty" jsonschema:"-"` // depends_on specifies which resources must be executed before the current one
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
	case "csv":
		return csv.New(rs.Spec)
	case "dockerdigest":
		return dockerdigest.New(rs.Spec)
	case "dockerfile":
		return dockerfile.New(rs.Spec)
	case "dockerimage":
		return dockerimage.New(rs.Spec)
	case "githubrelease":
		return githubrelease.New(rs.Spec)
	case "gitbranch":
		return gitbranch.New(rs.Spec)
	case "gittag":
		return gittag.New(rs.Spec)
	case "gitea/branch":
		return giteaBranch.New(rs.Spec)
	case "gitea/tag":
		return giteaTag.New(rs.Spec)
	case "gitea/release":
		return giteaRelease.New(rs.Spec)
	case "file":
		return file.New(rs.Spec)
	case "helmchart":
		return helm.New(rs.Spec)
	case "jenkins":
		return jenkins.New(rs.Spec)
	case "json":
		return json.New(rs.Spec)
	case "maven":
		return maven.New(rs.Spec)
	case "shell":
		return shell.New(rs.Spec)
	case "toml":
		return toml.New(rs.Spec)
	case "yaml":
		return yaml.New(rs.Spec)
	case "xml":
		return xml.New(rs.Spec)
	case "npm":
		return npm.New(rs.Spec)
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
		"csv":           &csv.Spec{},
		"dockerdigest":  &dockerdigest.Spec{},
		"dockerfile":    &dockerfile.Spec{},
		"dockerimage":   &dockerimage.Spec{},
		"file":          &file.Spec{},
		"gittag":        &gittag.Spec{},
		"gitbranch":     &gitbranch.Spec{},
		"gitea/branch":  &giteaBranch.Spec{},
		"gitea/release": &giteaRelease.Spec{},
		"gitea/tag":     &giteaTag.Spec{},
		"githubrelease": &githubrelease.Spec{},
		"helmchart":     &helm.Spec{},
		"jenkins":       &jenkins.Spec{},
		"json":          &json.Spec{},
		"maven":         &maven.Spec{},
		"shell":         &shell.Spec{},
		"toml":          &toml.Spec{},
		"xml":           &xml.Spec{},
		"yaml":          &yaml.Spec{},
		"npm":           &npm.Spec{},
	}
}

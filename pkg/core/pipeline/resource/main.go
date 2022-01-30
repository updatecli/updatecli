package resource

import (
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/transformer"
	"github.com/updatecli/updatecli/pkg/plugins/awsami"
	"github.com/updatecli/updatecli/pkg/plugins/dockerdigest"
	"github.com/updatecli/updatecli/pkg/plugins/dockerfile"
	"github.com/updatecli/updatecli/pkg/plugins/dockerimage"
	"github.com/updatecli/updatecli/pkg/plugins/file"
	"github.com/updatecli/updatecli/pkg/plugins/githubrelease"
	"github.com/updatecli/updatecli/pkg/plugins/gittag"
	"github.com/updatecli/updatecli/pkg/plugins/helm"
	"github.com/updatecli/updatecli/pkg/plugins/jenkins"
	"github.com/updatecli/updatecli/pkg/plugins/maven"
	"github.com/updatecli/updatecli/pkg/plugins/shell"
	"github.com/updatecli/updatecli/pkg/plugins/yaml"
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

// Unmarshal decode a source spec and returned its typed content
func (rs *ResourceConfig) Unmarshal() (resource Resource, err error) {
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
		return nil, fmt.Errorf("⚠ Don't support source kind: %v", rs.Kind)
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

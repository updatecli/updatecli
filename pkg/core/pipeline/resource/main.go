package resource

import (
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/transformer"
	"github.com/updatecli/updatecli/pkg/plugins/resources/awsami"
	"github.com/updatecli/updatecli/pkg/plugins/resources/bazelmod"
	"github.com/updatecli/updatecli/pkg/plugins/resources/cargopackage"
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
	gitlabBranch "github.com/updatecli/updatecli/pkg/plugins/resources/gitlab/branch"
	gitlabRelease "github.com/updatecli/updatecli/pkg/plugins/resources/gitlab/release"
	gitlabTag "github.com/updatecli/updatecli/pkg/plugins/resources/gitlab/tag"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gittag"
	"github.com/updatecli/updatecli/pkg/plugins/resources/go/gomod"
	golang "github.com/updatecli/updatecli/pkg/plugins/resources/go/language"
	gomodule "github.com/updatecli/updatecli/pkg/plugins/resources/go/module"
	"github.com/updatecli/updatecli/pkg/plugins/resources/hcl"
	"github.com/updatecli/updatecli/pkg/plugins/resources/helm"
	"github.com/updatecli/updatecli/pkg/plugins/resources/jenkins"
	"github.com/updatecli/updatecli/pkg/plugins/resources/json"
	"github.com/updatecli/updatecli/pkg/plugins/resources/maven"
	"github.com/updatecli/updatecli/pkg/plugins/resources/npm"
	"github.com/updatecli/updatecli/pkg/plugins/resources/shell"
	stashBranch "github.com/updatecli/updatecli/pkg/plugins/resources/stash/branch"
	stashTag "github.com/updatecli/updatecli/pkg/plugins/resources/stash/tag"
	"github.com/updatecli/updatecli/pkg/plugins/resources/temurin"
	terraformLock "github.com/updatecli/updatecli/pkg/plugins/resources/terraform/lock"
	terraformProvider "github.com/updatecli/updatecli/pkg/plugins/resources/terraform/provider"
	terraformRegistry "github.com/updatecli/updatecli/pkg/plugins/resources/terraform/registry"
	"github.com/updatecli/updatecli/pkg/plugins/resources/toml"
	"github.com/updatecli/updatecli/pkg/plugins/resources/toolversions"
	updateclihttp "github.com/updatecli/updatecli/pkg/plugins/resources/updateclihttp"
	"github.com/updatecli/updatecli/pkg/plugins/resources/xml"
	"github.com/updatecli/updatecli/pkg/plugins/resources/yaml"
)

type ResourceConfig struct {
	//"dependson" allows to specify the order of execution of resources.
	//It accepts a list of rules like "(resourceType#)resourceId(:booleanOperator)".
	//
	//The resourceType is optional and can be one of "condition", "source" or "target"
	//By default the resourceType is the current resource type
	//
	//The resourceId is the name of the resource to depend on
	//
	//The booleanOperator is optional and can be "AND" or "OR"
	//
	//examples:
	//    dependson:
	//      * condition#myCondition:and
	//      * source#mySource
	//
	//remarks:
	//  * The parameters "sourceid" and "conditionsids" affect the order of resource execution.
	//  * To avoid circular dependencies, the depended resource may need to remove any conditionids or set "disablesourceinput to true".
	DependsOn []string `yaml:",omitempty"`
	//name specifies the resource name
	Name string `yaml:",omitempty"`
	//kind specifies the resource kind which defines accepted spec value
	Kind string `yaml:",omitempty" jsonschema:"required"`
	//transformers defines how the default input value need to be transformed
	Transformers transformer.Transformers `yaml:",omitempty"`
	//spec specifies parameters for a specific resource kind
	Spec interface{} `yaml:",omitempty"`
	//scmid specifies the scm configuration key associated to the current resource
	SCMID string `yaml:",omitempty"` // SCMID references a uniq scm configuration
	//!deprecated, please use scmid
	//DeprecatedSCMID is kept for backward compatibility
	DeprecatedSCMID string `yaml:"scmID,omitempty" jsonschema:"-"`
	//!deprecated, please use dependson
	//DeprecatedDependsOn is kept for backward compatibility
	DeprecatedDependsOn []string `yaml:"depends_on,omitempty" jsonschema:"-"`
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

	case "bazelmod":

		return bazelmod.New(rs.Spec)

	case "cargopackage":

		return cargopackage.New(rs.Spec, rs.SCMID != "")

	case "csv":

		return csv.New(rs.Spec)

	case "dockerdigest":

		return dockerdigest.New(rs.Spec)

	case "dockerfile":

		return dockerfile.New(rs.Spec)

	case "dockerimage":

		return dockerimage.New(rs.Spec)

	case "gitbranch":

		return gitbranch.New(rs.Spec)

	case "gitea/branch":

		return giteaBranch.New(rs.Spec)

	case "gitea/tag":

		return giteaTag.New(rs.Spec)

	case "gitea/release":

		return giteaRelease.New(rs.Spec)

	case "githubrelease":

		return githubrelease.New(rs.Spec)

	case "gitlab/branch":

		return gitlabBranch.New(rs.Spec)

	case "gitlab/tag":

		return gitlabTag.New(rs.Spec)

	case "gitlab/release":

		return gitlabRelease.New(rs.Spec)

	case "gittag":

		return gittag.New(rs.Spec)

	case "golang":

		return golang.New(rs.Spec)

	case "golang/gomod":

		return gomod.New(rs.Spec)

	case "golang/module":

		return gomodule.New(rs.Spec)

	case "file":

		return file.New(rs.Spec)

	case "hcl":

		return hcl.New(rs.Spec)

	case "helmchart":

		return helm.New(rs.Spec)

	case "http":

		return updateclihttp.New(rs.Spec)

	case "jenkins":

		return jenkins.New(rs.Spec)

	case "json":

		return json.New(rs.Spec)

	case "maven":

		return maven.New(rs.Spec)

	case "npm":

		return npm.New(rs.Spec)

	case "shell":

		return shell.New(rs.Spec)

	case "stash/branch":

		return stashBranch.New(rs.Spec)

	case "stash/tag":

		return stashTag.New(rs.Spec)

	case "terraform/file":

		return hcl.New(rs.Spec)

	case "temurin":

		return temurin.New(rs.Spec)

	case "terraform/lock":

		return terraformLock.New(rs.Spec)

	case "terraform/provider":

		return terraformProvider.New(rs.Spec)

	case "terraform/registry":

		return terraformRegistry.New(rs.Spec)

	case "toml":

		return toml.New(rs.Spec)

	case "toolversions":

		return toolversions.New(rs.Spec)

	case "xml":

		return xml.New(rs.Spec)

	case "yaml":

		return yaml.New(rs.Spec)

	default:

		return nil, fmt.Errorf("%s Don't support resource kind: %v", result.FAILURE, rs.Kind)
	}
}

// Resource allow to manipulate a resource that can be a source, a condition or a target
type Resource interface {
	// Source returns the resource value
	Source(workingDir string, sourceResult *result.Source) error
	// Condition checks if the resource is in the expected state
	Condition(version string, scm scm.ScmHandler) (pass bool, message string, err error)
	// Target updates the resource with the given value
	Target(source string, scm scm.ScmHandler, dryRun bool, targetResult *result.Target) (err error)
	// Changelog returns the changelog for this resource, or an empty string if not supported
	Changelog(from, to string) *result.Changelogs
	// ReportConfig returns a new resource configuration
	// with only the necessary configuration fields without any sensitive information
	// or context specific data.
	ReportConfig() interface{}
}

// Need to do reflect of ResourceConfig
func GetResourceMapping() map[string]interface{} {
	return map[string]interface{}{
		"aws/ami":            &awsami.Spec{},
		"bazelmod":           &bazelmod.Spec{},
		"cargopackage":       &cargopackage.Spec{},
		"csv":                &csv.Spec{},
		"dockerdigest":       &dockerdigest.Spec{},
		"dockerfile":         &dockerfile.Spec{},
		"dockerimage":        &dockerimage.Spec{},
		"file":               &file.Spec{},
		"gittag":             &gittag.Spec{},
		"gitbranch":          &gitbranch.Spec{},
		"gitea/branch":       &giteaBranch.Spec{},
		"gitea/release":      &giteaRelease.Spec{},
		"gitea/tag":          &giteaTag.Spec{},
		"gitlab/branch":      &gitlabBranch.Spec{},
		"gitlab/release":     &gitlabRelease.Spec{},
		"gitlab/tag":         &gitlabTag.Spec{},
		"githubrelease":      &githubrelease.Spec{},
		"golang":             &golang.Spec{},
		"golang/gomod":       &gomod.Spec{},
		"golang/module":      &gomodule.Spec{},
		"hcl":                &hcl.Spec{},
		"helmchart":          &helm.Spec{},
		"http":               &updateclihttp.Spec{},
		"jenkins":            &jenkins.Spec{},
		"json":               &json.Spec{},
		"maven":              &maven.Spec{},
		"npm":                &npm.Spec{},
		"shell":              &shell.Spec{},
		"stash/branch":       &stashBranch.Spec{},
		"stash/tag":          &stashTag.Spec{},
		"temurin":            &temurin.Spec{},
		"terraform/file":     &hcl.Spec{},
		"terraform/lock":     &terraformLock.Spec{},
		"terraform/provider": &terraformProvider.Spec{},
		"terraform/registry": &terraformRegistry.Spec{},
		"toml":               &toml.Spec{},
		"toolversions":       &toolversions.Spec{},
		"xml":                &xml.Spec{},
		"yaml":               &yaml.Spec{},
	}
}

// GetReportConfig returns a clean version of the resource configuration
// without any sensitive information or context specific data.
func GetReportConfig(rs ResourceConfig) (any, error) {
	r, err := New(rs)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource %s: %w", rs.Kind, err)
	}

	newResourceConfig := rs
	newResourceConfig.Spec = r.ReportConfig()

	return newResourceConfig, nil
}

package resource

import (
	"fmt"
	"strings"

	"github.com/iancoleman/orderedmap"
	"github.com/invopop/jsonschema"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config/jsonschema/commentmap"
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
	// Name defines a resource name
	Name string
	// Kind defines a resource kind such as yaml
	Kind string
	// **Deprecated** Please consider `Transformers`
	Prefix string
	// **Deprecated** Please consider `Transformers`
	Postfix string
	// Define a lit of Transformer
	Transformers transformer.Transformers
	// Define resource spec according its kind
	Spec interface{} `jsonschema:"type=object"`
	// **Deprecated** Please look in the top scms resource
	Scm map[string]interface{}
	// Define which scm resource is linked to the current resource
	SCMID string `yaml:"scmID"` // SCMID references a uniq scm configuration
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

// Need to do reflect of ResourceConfig
func (ResourceConfig) JSONSchema() *jsonschema.Schema {

	type Resource struct {
		Kind string
		Spec interface{}
	}

	type ResourceConfigSchema ResourceConfig

	type Resources []Resource

	// SchemaResources contains list of every resource mapping
	var SchemaResources Resources = Resources{
		Resource{
			Kind: "jenkins",
			Spec: &jenkins.Spec{},
		},
		Resource{
			Kind: "file",
			Spec: &file.Spec{},
		},
		Resource{
			Kind: "yaml",
			Spec: &yaml.Spec{},
		},
		Resource{
			Kind: "aws/ami",
			Spec: &awsami.Spec{},
		},
	}

	var err error

	r := new(jsonschema.Reflector)

	r.Anonymous = true
	r.YAMLEmbeddedStructs = true
	r.DoNotReference = true
	r.RequiredFromJSONSchemaTags = true
	r.KeyNamer = strings.ToLower

	commentMap, err := commentmap.Get("../../../../pkg")

	if err != nil {
		logrus.Errorf(err.Error())
		return nil
	}

	resourceSchema := r.Reflect(ResourceConfigSchema{})

	for _, updatecliResource := range SchemaResources {
		r := new(jsonschema.Reflector)

		r.Anonymous = true
		r.PreferYAMLSchema = true
		r.YAMLEmbeddedStructs = true
		r.DoNotReference = true
		r.RequiredFromJSONSchemaTags = true
		r.CommentMap = commentMap
		r.KeyNamer = strings.ToLower

		// schema we need a way to remove schema

		// Main resource type schema such as source or condition
		s := r.Reflect(updatecliResource.Spec)
		s.ContentSchema = nil

		var spec jsonschema.Schema

		spec.Type = "object"
		spec.Required = []string{"kind"}
		spec.Properties = orderedmap.New()
		spec.Properties.Set("spec", s.Properties)
		spec.Properties.Set("kind", jsonschema.Schema{
			Enum: []interface{}{updatecliResource.Kind}})

		resourceSchema.OneOf = append(resourceSchema.OneOf, &spec)

	}

	return resourceSchema
}

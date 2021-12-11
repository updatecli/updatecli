package resource

import "github.com/updatecli/updatecli/pkg/core/transformer"

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

// type ResourceSpec interface {
// }

// Resource allow to manipulate a resource that can be a source, a condition or a target
type Resource interface {
	// Initialize a new Resource from its specification
	New() (*Resource, error)
	// // Validate the Resource or throws an error (with the validation issues)
	// Validate() error
	// // Returns the resource specification
	// Spec() ResourceSpec
	// Source() source.Source
	// Condition() condition.Condition
	// Target() target.Target
}

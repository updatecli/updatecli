package registry

import (
	"errors"
	"fmt"
	"slices"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

/*
"terraform/registry" defines the specification for retrieving Terraform provider or module versions from a Terraform registry.
It can be used as a "source" or a "condition".
*/
type Spec struct {
	// "type" defines the type of registry object to look up.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * "type" is required.
	//   * accepted values are "module" and "provider".
	//
	// example:
	//   * type: provider
	//
	Type string `yaml:",omitempty"`
	// "hostname" defines the hostname of the registry hosting the provider or module.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   registry.terraform.io
	//
	// remark:
	//   * "hostname" and "rawstring" are mutually exclusive.
	//   * applies to modules and providers.
	//
	// example:
	//   * hostname: app.terraform.io
	//
	Hostname string `yaml:",omitempty"`
	// "namespace" defines the namespace of the provider or module.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * required unless "rawstring" is set.
	//   * "namespace" and "rawstring" are mutually exclusive.
	//   * applies to modules and providers.
	//
	// example:
	//   * namespace: hashicorp
	//
	Namespace string `yaml:",omitempty"`
	// "name" defines the name of the provider or module.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * required unless "rawstring" is set.
	//   * "name" and "rawstring" are mutually exclusive.
	//   * applies to modules and providers.
	//
	// example:
	//   * name: kubernetes
	//
	Name string `yaml:",omitempty"`
	// "targetsystem" defines the target system of the module in the registry.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * required for type "module" unless "rawstring" is set.
	//   * "targetsystem" and "rawstring" are mutually exclusive.
	//   * only applies to modules.
	//
	// example:
	//   * targetsystem: aws
	//
	TargetSystem string `yaml:",omitempty"`
	// "rawstring" defines the provider or module reference in the registry as a single string.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * applies to modules and providers.
	//   * "rawstring" is mutually exclusive with "hostname", "namespace", "name" and "targetsystem".
	//
	// example:
	//   * rawstring: hashicorp/kubernetes
	//   * rawstring: registry.terraform.io/hashicorp/kubernetes
	//   * rawstring: terraform-aws-modules/vpc/aws
	//   * rawstring: app.terraform.io/terraform-aws-modules/vpc/aws
	//
	RawString string `yaml:",omitempty"`
	// "version" defines the version to check.
	//
	// compatible:
	//   * condition
	//
	// default:
	//   the output of the associated source.
	//
	Version string `yaml:",omitempty"`
	// "versionfilter" defines the filter used to select the version, such as a regex, semver or latest pattern.
	//
	// compatible:
	//   * source
	//
	// default:
	//   kind: semver
	//   pattern: "*"
	//
	VersionFilter version.Filter `yaml:",omitempty"`
}

const (
	TypeProvider = "provider"
	TypeModule   = "module"
)

var (
	AllowedTypes = []string{TypeProvider, TypeModule}
	// ErrSpecTypeUndefined is returned if a type wasn't specified
	ErrSpecTypeUndefined = errors.New("terraform/registry type undefined")
	// ErrSpecNamespaceUndefined is returned if a namespace wasn't specified
	ErrSpecNamespaceUndefined = errors.New("terraform/registry namespace undefined")
	// ErrSpecNameUndefined is returned if a name wasn't specified
	ErrSpecNameUndefined = errors.New("terraform/registry name undefined")
	// ErrSpecTargetSystemUndefined is returned if a targetsystem wasn't specified
	ErrSpecTargetSystemUndefined = errors.New("terraform/registry targetsystem undefined")
	// ErrSpecTypeNotAllowed is returned if a type wasn't allowed
	ErrSpecTypeNotAllowed = fmt.Errorf("terraform/registry type must be one of: %v", AllowedTypes)
	// ErrSpecRawStringAndHostnameDefined when we both spec RawString and Hostname, Namespace, Name, or TargetSystem have been specified
	ErrSpecRawStringAndHostnameDefined = errors.New("terraform/registry rawstring and hostname are mutually exclusive")
	// ErrSpecRawStringAndNamespaceDefined when we both spec RawString and Hostname, Namespace, Name, or TargetSystem have been specified
	ErrSpecRawStringAndNamespaceDefined = errors.New("terraform/registry rawstring and namespace are mutually exclusive")
	// ErrSpecRawStringAndNameDefined when we both spec RawString and Hostname, Namespace, Name, or TargetSystem have been specified
	ErrSpecRawStringAndNameDefined = errors.New("terraform/registry rawstring and name are mutually exclusive")
	// ErrSpecRawStringAndTargetSystemDefined when we both spec RawString and Hostname, Namespace, Name, or TargetSystem have been specified
	ErrSpecRawStringAndTargetSystemDefined = errors.New("terraform/registry rawstring and targetsystem are mutually exclusive")
	// ErrSpecProviderTargetSystemDefined is returned if a type wasn't specified
	ErrSpecProviderTargetSystemDefined = errors.New("terraform/registry type provider does not support targetsystem")
	// ErrWrongSpec is returned when the Spec has wrong content
	ErrWrongSpec error = errors.New("wrong spec content")
)

func (s *Spec) Validate() error {
	var errs []error

	if len(s.Type) == 0 {
		errs = append(errs, ErrSpecTypeUndefined)
	}

	if !slices.Contains(AllowedTypes, s.Type) {
		errs = append(errs, ErrSpecTypeNotAllowed)
	}

	if len(s.RawString) > 0 && len(s.Hostname) > 0 {
		errs = append(errs, ErrSpecRawStringAndHostnameDefined)
	}

	if len(s.RawString) > 0 && len(s.Namespace) > 0 {
		errs = append(errs, ErrSpecRawStringAndNamespaceDefined)
	}

	if len(s.RawString) > 0 && len(s.Name) > 0 {
		errs = append(errs, ErrSpecRawStringAndNameDefined)
	}

	if len(s.RawString) > 0 && len(s.TargetSystem) > 0 {
		errs = append(errs, ErrSpecRawStringAndTargetSystemDefined)
	}

	if len(s.RawString) == 0 && (len(s.Namespace) == 0) {
		errs = append(errs, ErrSpecNamespaceUndefined)
	}

	if len(s.RawString) == 0 && (len(s.Name) == 0) {
		errs = append(errs, ErrSpecNameUndefined)
	}

	if s.Type == TypeProvider {
		if len(s.TargetSystem) > 0 {
			errs = append(errs, ErrSpecProviderTargetSystemDefined)
		}
	}

	if s.Type == TypeModule {
		if len(s.RawString) == 0 && (len(s.TargetSystem) == 0) {
			errs = append(errs, ErrSpecTargetSystemUndefined)
		}
	}

	for _, e := range errs {
		logrus.Errorln(e)
	}

	if len(errs) > 0 {
		return ErrWrongSpec
	}

	return nil
}

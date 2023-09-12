package registry

import (
	"errors"
	"fmt"
	"slices"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

type Spec struct {
	/*
		"type" defines the type registry request to look up.

		compatible:
			* source
			* condition

		Supported values: module, provider
	*/
	Type string `yaml:",omitempty"`
	/*
		"hostname" the hostname of the provider or module.

		compatible:
			* source
			* condition

		remark:
			* Optional
			* Not allowed with rawstring.
			* Applicable for module and provider.
	*/
	Hostname string `yaml:",omitempty"`
	/*
		"namespace" the namespace of the provider or module

		compatible:
			* source
			* condition

		remark:
			* Required unless using rawstring
			* Not allowed with rawstring.
			* Applicable for module and provider.
	*/
	Namespace string `yaml:",omitempty"`
	/*
		"name" the name of the provider or module.

		compatible:
			* source
			* condition

		remark:
			* Required unless using rawstring
			* Not allowed with rawstring.
			* Applicable for module and provider.
	*/
	Name string `yaml:",omitempty"`
	/*
		"targetsystem" the target system for the module in registry

		compatible:
			* source
			* condition

		remark:
			* Required for type module unless using rawstring
			* Not allowed with rawstring
			* Applicable for module.
	*/
	TargetSystem string `yaml:",omitempty"`

	/*
		"rawstring" provider reference to registry in single string.

		compatible:
			* source
			* condition

		Examples:
			* hashicorp/kubernetes
			* registry.terraform.io/hashicorp/kubernetes
			* terraform-aws-modules/vpc/aws
			* app.terraform.io/terraform-aws-modules/vpc/aws

		remark:
			* Applicable for module and provider.
			* Not allowed with hostname, namespace, name, and targetsystem.
	*/
	RawString string `yaml:",omitempty"`

	/*
		"version" defines a specific version to be used during condition check.

		compatible:
			* condition
	*/
	Version string `yaml:",omitempty"`
	/*
		"versionfilter" provides parameters to specify version pattern and its type like regex, semver, or just latest.

		compatible:
			* source
	*/
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
			errs = append(errs, ErrSpecProviderTargetSystemDefined)
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

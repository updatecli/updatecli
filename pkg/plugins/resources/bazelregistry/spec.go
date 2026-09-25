package bazelregistry

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

/*
"bazelregistry" defines the specification for retrieving a Bazel module version from a Bazel registry.
It can be used as a "source" or a "condition".
*/
type Spec struct {
	// "module" defines the name of the Bazel module to query from the registry.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * it is required.
	//
	// example:
	//   * module: rules_go
	//   * module: rules_python
	//   * module: gazelle
	//
	Module string `yaml:",omitempty" jsonschema:"required"`
	// "versionfilter" defines the version pattern and its kind, such as regex, semver or latest.
	//
	// compatible:
	//   * source
	//
	// default:
	//   kind: latest
	//
	// remark:
	//   * yanked versions are ignored.
	//   * with the default filter, the highest semantic version is returned.
	//
	VersionFilter version.Filter `yaml:",omitempty"`
	// "url" defines a custom registry URL.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   https://raw.githubusercontent.com/bazelbuild/bazel-central-registry/main/modules/{module}/metadata.json
	//
	// remark:
	//   * the URL must contain the "{module}" placeholder, which is replaced with the module name.
	//   * when unset, the official Bazel Central Registry is used.
	//
	// example:
	//   * url: https://raw.githubusercontent.com/bazelbuild/bazel-central-registry/main/modules/{module}/metadata.json
	//   * url: https://mycompany.com/bazel-registry/modules/{module}/metadata.json
	//
	URL string `yaml:",omitempty"`
}

var (
	// ErrSpecModuleUndefined is returned if a module wasn't specified
	ErrSpecModuleUndefined = errors.New("bazelregistry module undefined")
	// ErrWrongSpec is returned when the Spec has wrong content
	ErrWrongSpec = errors.New("wrong spec content")
)

const (
	// DefaultRegistryURL is the default Bazel Central Registry URL template
	DefaultRegistryURL = "https://raw.githubusercontent.com/bazelbuild/bazel-central-registry/main/modules/{module}/metadata.json"
)

// Validate tests that the spec has the required fields
func (s *Spec) Validate() error {
	var errs []error

	if len(s.Module) == 0 {
		errs = append(errs, ErrSpecModuleUndefined)
	}

	for _, e := range errs {
		logrus.Errorln(e)
	}

	if len(errs) > 0 {
		return ErrWrongSpec
	}

	return nil
}

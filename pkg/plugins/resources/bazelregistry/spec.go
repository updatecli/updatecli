package bazelregistry

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines a specification for a "bazelregistry" resource
// parsed from an updatecli manifest file
type Spec struct {
	// Module specifies the Bazel module name to query from the registry
	//
	// compatible:
	//   * source
	//   * condition
	//
	// example:
	//   * rules_go
	//   * rules_python
	//   * gazelle
	Module string `yaml:",omitempty" jsonschema:"required"`
	// VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	//
	// compatible:
	//   * source
	//
	// default:
	//   kind: latest
	VersionFilter version.Filter `yaml:",omitempty"`
	// URL specifies the custom registry URL (defaults to Bazel Central Registry)
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   https://raw.githubusercontent.com/bazelbuild/bazel-central-registry/main/modules/{module}/metadata.json
	//
	// example:
	//   * https://raw.githubusercontent.com/bazelbuild/bazel-central-registry/main/modules/{module}/metadata.json
	//   * https://mycompany.com/bazel-registry/modules/{module}/metadata.json
	//
	// remarks:
	//   * The URL must contain {module} placeholder which will be replaced with the module name
	//   * If not specified, defaults to the official Bazel Central Registry
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

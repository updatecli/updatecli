package toml

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

/*
"toml" defines the specification for manipulating toml files.
It can be used as a "source", a "condition", or a "target".
*/
type Spec struct {
	// "engine" defines the engine used to manipulate the toml file.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// default:
	//   dasel/v1
	//
	// remark:
	//   * accepted values are "dasel/v1", "dasel/v2", "dasel/v3" and "dasel".
	//   * "dasel" selects the latest dasel engine, currently "dasel/v3".
	//   * "dasel/v1" and "dasel/v2" are deprecated in favor of "dasel/v3".
	//
	// example:
	//   * engine: dasel/v3
	//
	Engine *string `yaml:",omitempty"`
	// "file" defines the path of the toml file to use.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * "file" and "files" are mutually exclusive.
	//   * the schemes "https://", "http://" and "file://" are supported in a source or a condition.
	//
	// example:
	//   * file: Cargo.toml
	//
	File string `yaml:",omitempty"`
	// "files" defines the list of toml file paths to use.
	//
	// compatible:
	//   * condition
	//   * target
	//
	// remark:
	//   * "file" and "files" are mutually exclusive.
	//   * the schemes "https://", "http://" and "file://" are supported in a condition.
	//
	Files []string `yaml:",omitempty"`
	// "query" defines an advanced dasel v1 query returning several values.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * "key" or "query" is required.
	//   * "query" is only used by the "dasel/v1" engine. The engines "dasel/v2"
	//     and "dasel/v3" require "key" instead.
	//   * with the "dasel/v1" engine, "query" takes precedence over "key".
	//   * in a source, "query" and "versionfilter" must be used together.
	//   * "query" accepts a dasel query, more information on https://github.com/tomwright/dasel
	//
	Query string `yaml:",omitempty"`
	// "key" defines the toml key path to use.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * "key" or "query" is required.
	//   * "key" accepts a dasel query matching the selected engine,
	//     more information on https://github.com/tomwright/dasel
	//
	// example:
	//   * key: package.version
	//
	Key string `yaml:",omitempty"`
	// "value" defines the value associated with the toml key.
	//
	// compatible:
	//   * condition
	//   * target
	//
	// default:
	//   the output of the associated source.
	//
	Value string `yaml:",omitempty"`
	// "multiple" retrieves several values from a query.
	//
	// compatible:
	//   * condition
	//   * target
	//
	// deprecated:
	//   * use "query" instead. A "key" combined with "multiple" is converted to "query".
	//
	Multiple bool `yaml:",omitempty" jsonschema:"-"`
	// "versionfilter" defines the version pattern and its kind, such as "regex", "semver" or "latest".
	//
	// compatible:
	//   * source
	//
	// remark:
	//   * with the "dasel/v1" engine, "versionfilter" and "query" must be used together.
	//   * more information on https://www.updatecli.io/docs/core/versionfilter/
	//
	VersionFilter version.Filter `yaml:",omitempty"`
	// "createmissingkey" creates the key when the toml file does not hold it yet.
	//
	// compatible:
	//   * target
	//
	// default:
	//   false
	//
	// remark:
	//   * when false, a missing key raises an error.
	//   * only supported with "key".
	//   * not supported by the "dasel/v3" engine, which cannot create missing keys.
	//
	CreateMissingKey bool `yaml:",omitempty"`
}

var (
	// ErrSpecFileUndefined is returned if a file wasn't specified
	ErrSpecFileUndefined = errors.New("toml file undefined")
	// ErrSpecKeyUndefined is returned if a key wasn't specified
	ErrSpecKeyUndefined = errors.New("toml key or query undefined")
	// ErrSpecFileAndFilesDefines when we both spec File and Files have been specified
	ErrSpecFileAndFilesDefined = errors.New("parameter \"file\" and \"files\" are mutually exclusive")
	// ErrWrongSpec is returned when the Spec has wrong content
	ErrWrongSpec error = errors.New("wrong spec content")
)

const (
	ENGINEDASEL_V1 = "dasel/v1"
	ENGINEDASEL_V2 = "dasel/v2"
	ENGINEDASEL_V3 = "dasel/v3"
	// ENGINEDASEL_LATEST is an alias resolving to the latest dasel engine.
	ENGINEDASEL_LATEST = "dasel"
	ENGINEDEFAULT      = ENGINEDASEL_V1
)

// resolveEngine normalizes a user-provided engine value, resolving the "dasel"
// alias to the latest supported engine. An empty value resolves to the default.
func resolveEngine(engine string) string {
	switch engine {
	case "":
		return ENGINEDEFAULT
	case ENGINEDASEL_LATEST:
		return ENGINEDASEL_V3
	default:
		return engine
	}
}

func (s *Spec) Validate() error {
	var errs []error

	if len(s.File) == 0 && len(s.Files) == 0 {
		errs = append(errs, ErrSpecFileUndefined)
	}
	if len(s.Key) == 0 && len(s.Query) == 0 {
		errs = append(errs, ErrSpecKeyUndefined)
	}

	if len(s.File) > 0 && len(s.Files) > 0 {
		errs = append(errs, ErrSpecFileAndFilesDefined)
	}

	engine := ENGINEDEFAULT
	if s.Engine != nil {
		engine = resolveEngine(*s.Engine)
	}

	// dasel v2 and v3 deprecate the "query" parameter in favor of "key".
	if (engine == ENGINEDASEL_V2 || engine == ENGINEDASEL_V3) && len(s.Query) > 0 && len(s.Key) == 0 {
		errs = append(errs, fmt.Errorf("engine %q requires the parameter \"key\" over \"query\"", engine))
	}

	// The dasel v3 engine cannot create missing keys (its API resolves the key
	// before setting a value), so "createmissingkey" is incompatible with it.
	if engine == ENGINEDASEL_V3 && s.CreateMissingKey {
		errs = append(errs, fmt.Errorf("engine %q does not support the parameter \"createmissingkey\"", engine))
	}

	for _, e := range errs {
		logrus.Errorln(e)
	}

	if len(errs) > 0 {
		return ErrWrongSpec
	}

	return nil
}

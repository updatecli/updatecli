package csv

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

/*
"csv" defines the specification for manipulating csv files.
It can be used as a "source", a "condition", or a "target".
*/
type Spec struct {
	// "engine" defines the engine used to manipulate the csv file.
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
	//   * "dasel" is the latest dasel engine, currently dasel v3.
	//   * "dasel/v1" and "dasel/v2" are deprecated in favour of "dasel/v3".
	//
	// example:
	//   * engine: dasel/v3
	//
	Engine *string `yaml:",omitempty"`
	// "file" defines the path of the csv file.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * "file" and "files" are mutually exclusive.
	//   * a "file://" prefix is removed.
	//   * the schemes "https://" and "http://" are not supported in a target.
	//
	File string `yaml:",omitempty"`
	// "files" defines the list of csv file paths.
	//
	// compatible:
	//   * condition
	//   * target
	//
	// remark:
	//   * "file" and "files" are mutually exclusive.
	//
	Files []string `yaml:",omitempty"`
	// "key" defines the csv query.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * "key" or "query" is required.
	//   * the engines "dasel/v2" and "dasel/v3" require "key" instead of "query".
	//
	Key string `yaml:",omitempty"`
	// "query" defines an advanced csv query, returning several results.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * it overrides "key".
	//   * only supported by the engine "dasel/v1".
	//   * with the engine "dasel/v1", "query" and "versionfilter" must be used together in a source.
	//
	Query string `yaml:",omitempty"`
	// "value" defines the csv value.
	//
	// compatible:
	//   * condition
	//   * target
	//
	// default:
	//   the output of the associated source.
	//
	Value string `yaml:",omitempty"`
	// "comma" defines the csv separator character.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// default:
	//   ,
	//
	Comma rune `yaml:",omitempty"`
	// "comment" defines the csv comment character.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// default:
	//   #
	//
	Comment rune `yaml:",omitempty"`
	// "multiple" enables queries returning several results.
	//
	// compatible:
	//   * condition
	//   * target
	//
	// deprecated:
	//   * use "query" instead.
	//
	Multiple bool `yaml:",omitempty" jsonschema:"-"`
	// "versionfilter" defines the version pattern and its kind, such as regex, semver or latest.
	//
	// compatible:
	//   * source
	//
	// default:
	//   kind: latest
	//
	// remark:
	//   * with the engine "dasel/v1", "query" and "versionfilter" must be used together.
	//
	VersionFilter version.Filter `yaml:",omitempty"`
}

var (
	ErrSpecFileUndefined       = errors.New("csv file undefined")
	ErrSpecKeyUndefined        = errors.New("csv key or query undefined")
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

	if len(errs) > 0 {
		for i := range errs {
			logrus.Errorln(errs[i])
		}
		return ErrWrongSpec
	}

	return nil
}

package json

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

/*
"json"  defines the specification for manipulating "json" files.
*/
type Spec struct {
	// engine defines the engine used to manipulate the json file.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target

	// default:
	//   * "dasel/v1" is the default engine used to manipulate json files
	//
	// accepted values:
	//   * "dasel/v1" for dasel v1 engine
	//   * "dasel/v2" for dasel v2 engine
	//   * "dasel/v3" for dasel v3 engine
	//   * "dasel" for the latest dasel engine which is currently dasel v3
	Engine *string `yaml:",omitempty"`
	// file defines the Json file path to interact with.
	//
	// compatible:
	//    * source
	//    * condition
	//    * target

	// 	remark:
	//    * "file" and "files" are mutually exclusive
	//    * scheme "https://", "http://", and "file://" are supported in path for source and condition
	File string `yaml:",omitempty"`
	// files defines the list of Json files path to interact with.
	//
	// compatible:
	//   * condition
	//   * target

	// remark:
	//   * "file" and "files" are mutually exclusive
	//   * scheme "https://", "http://", and "file://" are supported in path for source and condition
	Files []string `yaml:",omitempty"`
	// key defines the Jsonpath key to manipulate.
	//
	// compatible:
	//  * source
	// 	* condition
	// 	* target
	//
	// remark:
	// 	* key is a simpler version of Jsonpath accepts keys.
	// 	* key accepts Dasel query, more information on https://github.com/tomwright/dasel
	//  * key accepts values based on the engine used
	//
	// example:
	// 	* key: $.name
	// 	* key: name
	// 	* file: https://nodejs.org/dist/index.json
	// 	  key: .(lts!=false).version
	Key string `yaml:",omitempty"`
	// value defines the Jsonpath key value to manipulate. Default to source output.
	//
	// compatible:
	//  * condition
	// 	* target
	//
	// default:
	// 	when used for a condition or a target, the default value is the output of the source.
	Value string `yaml:",omitempty"`
	// query defines the Jsonpath query to manipulate. It accepts advanced Dasel v1 query
	// this parameter is now deprecated in Dasel v2 and replaced by the parameter "key".
	//
	// compatible:
	// 	* source
	// 	* condition
	// 	* target
	//
	// example:
	// 	* query: .name
	// 	* query: ".[*].tag_name"
	//
	// remark:
	// 	* query accepts Dasel query, more information on https://github.com/tomwright/dasel
	Query string `yaml:",omitempty"`
	// versionfilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	// more information on https://www.updatecli.io/docs/core/versionfilter/
	//
	// compatible:
	// 	* source

	VersionFilter version.Filter `yaml:",omitempty"`
	// multiple allows to retrieve multiple values from a query. *Deprecated* Please look at query parameter to achieve similar objective
	//
	// compatible:
	// 	* condition
	// 	* target
	Multiple bool `yaml:",omitempty" jsonschema:"-"`
}

var (
	// ErrSpecFileUndefined is returned when the Spec has no file defined
	ErrSpecFileUndefined = errors.New("json file undefined")
	// ErrSpecKeyUndefined is returned when the Spec has no key defined
	ErrSpecKeyUndefined = errors.New("json key or query undefined")
	// ErrSpecFileAndFilesDefined is returned when the Spec has both file and files defined
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

// Validate checks if the Spec is properly defined
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

	for _, e := range errs {
		logrus.Errorln(e)
	}

	if len(errs) > 0 {
		return ErrWrongSpec
	}

	return nil
}

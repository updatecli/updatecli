package gomod

import (
	"errors"
)

var (
	// ErrWrongSpec is returned when the Spec has wrong content
	ErrWrongSpec error = errors.New("wrong spec content")
)

/*
"golang/gomod" defines the specification for manipulating a go.mod file.
It can be used as a "source", a "condition", or a "target".
*/
type Spec struct {
	// "file" defines the path of the go.mod file to use.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// default:
	//   go.mod
	//
	// remark:
	//   * the schemes "https://", "http://" and "file://" are supported in a source or a condition.
	//
	// example:
	//   * file: go.mod
	//   * file: tools/go.mod
	//
	File string `yaml:",omitempty"`
	// "module" defines the path of the Go module to manipulate.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// default:
	//   empty
	//
	// remark:
	//   * when empty, the Go version set by the "go" directive is used instead of a module version.
	//
	// example:
	//   * module: github.com/sirupsen/logrus
	//
	Module string `yaml:",omitempty"`
	// "indirect" defines whether the module is an indirect dependency.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// default:
	//   false
	//
	// remark:
	//   * when false, only direct dependencies match. When true, only indirect dependencies match.
	//   * ignored when "replace" is true.
	//
	Indirect bool `yaml:",omitempty"`
	// "version" defines the version to check or to set.
	//
	// compatible:
	//   * condition
	//   * target
	//
	// default:
	//   the output of the associated source.
	//
	// remark:
	//   * it is the module version when "module" is set, otherwise the Go version.
	//
	// example:
	//   * version: v1.9.3
	//   * version: 1.23.0
	//
	Version string `yaml:",omitempty"`
	// "replace" defines whether to manipulate the module of a replace directive instead of a require directive.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// default:
	//   false
	//
	// remark:
	//   * the version retrieved or updated is the one of the replacement module, on the right hand side.
	//
	Replace bool `yaml:",omitempty"`
	// "replaceversion" defines the version of the replaced module, on the left hand side of the replace directive.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// default:
	//   empty, which matches any version of the replaced module.
	//
	// remark:
	//   * only used when "replace" is true.
	//   * for the replace directive `moduleA v1.2.3 => moduleB v1.0.0`, set "module" to "moduleA"
	//     and "replaceversion" to "v1.2.3".
	//
	// example:
	//   * replaceversion: v1.2.3
	//
	ReplaceVersion string `yaml:",omitempty"`
}

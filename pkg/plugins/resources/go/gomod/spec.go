package gomod

import (
	"errors"
)

var (
	// ErrWrongSpec is returned when the Spec has wrong content
	ErrWrongSpec error = errors.New("wrong spec content")
)

// Spec defines a specification for a "Golang" resource parsed from an updatecli manifest file
type Spec struct {
	// File defines the go.mod file, default to "go.mod"
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//  * scheme "https://", "http://", and "file://" are supported in path for source and condition
	//
	File string `yaml:",omitempty"`
	// Module defines the module path
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//  * scheme "https://", "http://", and "file://" are supported in path for source and condition
	//
	Module string `yaml:",omitempty"`
	// Indirect specifies if we manipulate an indirect dependency
	//
	// compatible:
	//   * source
	//   * condition
	//
	Indirect bool `yaml:",omitempty"`
	// Version Defines a specific golang version
	//
	// compatible:
	//   * source
	//   * condition
	//
	Version string `yaml:",omitempty"`
	// Replace specifies if we manipulate a replaced dependency
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	Replace bool `yaml:",omitempty"`
	// ReplaceVersion specifies the specific Go module version to replace
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// default: unset, which will match any version of the module being replaced.
	//
	// Example:
	//  For the following Go replace instruction:
	//    moduleA v1.2.3 => moduleB v1.0.0
	//  - The 'module' field should be set to 'moduleA' (the module being replaced, left-hand side).
	//  - The value of ReplaceVersion would be 'v1.0.0', corresponding to the version of moduleB
	//    (the module used as replacement, right-hand side).
	ReplaceVersion string `yaml:",omitempty"`
}

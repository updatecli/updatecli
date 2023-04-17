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
	File string `yaml:",omitempty"`
	// Module defines the module path
	Module string `yaml:",omitempty"`
	// Indirect specifies if we manipulate an indirect dependency
	Indirect bool `yaml:",omitempty"`
	// Version Defines a specific golang version
	Version string `yaml:",omitempty"`
}

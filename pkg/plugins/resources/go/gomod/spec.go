package gomod

import (
	"errors"

	"github.com/sirupsen/logrus"
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
	// Kind defines the kind of information we are manipulating, accepted value are "golang" or "module"
	Kind string `yaml:",omitempty"`
}

func (s Spec) Validate() error {
	if s.Kind != "" &&
		s.Kind != kindGolang &&
		s.Kind != kindModule {
		logrus.Errorf("wrong kind %q, accepted value %v",
			s.Kind, []string{"", kindGolang, kindModule})
		return ErrWrongSpec
	}
	return nil
}

package toml

import (
	"errors"

	"github.com/sirupsen/logrus"
)

type Spec struct {
	// [s][c][t] File specifies the toml file to manipulate
	File string `yaml:",omitempty"`
	// [c][t] Files specifies a list of Json file to manipuate
	Files []string `yaml:",omitempty"`
	// [s][c][t] Key specifies the query to retrieve an information from a toml file
	Key string `yaml:",omitempty"`
	// [s][c][t] Value specifies the value for a specific key. Default to source output
	Value string `yaml:",omitempty"`
	// [c][t] Multiple allows to query multiple values at once
	Multiple bool `yaml:",omitempty"`
}

var (
	// ErrSpecFileUndefined is returned if a file wasn't specified
	ErrSpecFileUndefined = errors.New("toml file undefined")
	// ErrSpecKeyUndefined is returned if a key wasn't specified
	ErrSpecKeyUndefined = errors.New("toml key undefined")
	// ErrSpecFileAndFilesDefines when we both spec File and Files have been specified
	ErrSpecFileAndFilesDefined = errors.New("parameter \"file\" and \"files\" are mutually exclusive")
	// ErrWrongSpec is returned when the Spec has wrong content
	ErrWrongSpec error = errors.New("wrong spec content")
)

func (s *Spec) Validate() error {
	var errs []error

	if len(s.File) == 0 && len(s.Files) == 0 {
		errs = append(errs, ErrSpecFileUndefined)
	}
	if len(s.Key) == 0 {
		errs = append(errs, ErrSpecKeyUndefined)
	}

	if len(s.File) > 0 && len(s.Files) > 0 {
		errs = append(errs, ErrSpecFileAndFilesDefined)
	}

	for _, e := range errs {
		logrus.Errorln(e)
	}

	if len(errs) > 0 {
		return ErrWrongSpec
	}

	return nil
}

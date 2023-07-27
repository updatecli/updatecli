package hcl

import (
	"errors"

	"github.com/sirupsen/logrus"
)

type Spec struct {
	// [s][c][t] File specifies the hcl file to manipulate
	File string `yaml:",omitempty"`
	// [s][c][t] Files specifies a list of hcl file to manipulate
	Files []string `yaml:",omitempty"`
	// [s][c][t] Key specifies the query to retrieve information from a hcl file
	Key string `yaml:",omitempty"`
	// [c][t] Value specifies the value for a specific key. Default to source output
	Value string `yaml:",omitempty"`
}

var (
	// ErrSpecFileUndefined is returned if a file wasn't specified
	ErrSpecFileUndefined = errors.New("hcl file undefined")
	// ErrSpecKeyUndefined is returned if a key wasn't specified
	ErrSpecKeyUndefined = errors.New("hcl key undefined")
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

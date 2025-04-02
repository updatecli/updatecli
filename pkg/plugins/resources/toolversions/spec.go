package toolversions

import (
	"errors"

	"github.com/sirupsen/logrus"
)

type Spec struct {
	// [s][c][t] File specifies the .tool-versions file to manipulate
	File string `yaml:",omitempty"`
	// [c][t] Files specifies a list of .tool-versions file to manipulate
	Files []string `yaml:",omitempty"`
	// [s][c][t] Key specifies the query to retrieve an information from a .tool-versions file
	Key string `yaml:",omitempty"`
	// [s][c][t] Value specifies the value for a specific key. Default to source output
	Value string `yaml:",omitempty"`
	/*
	  [t] CreateMissingKey allows non-existing keys. If the key does not exist, the key is created if AllowsMissingKey
	  is true, otherwise an error is raised (the default).
	  Only supported if Key is used
	*/
	CreateMissingKey bool `yaml:",omitempty"`
}

var (
	// ErrSpecFileUndefined is returned if a file wasn't specified
	ErrSpecFileUndefined = errors.New(".tool-versions file undefined")
	// ErrSpecKeyUndefined is returned if a key wasn't specified
	ErrSpecKeyUndefined = errors.New("tool-versions key undefined")
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

	if len(s.File) > 0 && len(s.Files) > 0 {
		errs = append(errs, ErrSpecFileAndFilesDefined)
	}

	if len(s.Key) == 0 {
		errs = append(errs, ErrSpecKeyUndefined)
	}

	for _, e := range errs {
		logrus.Errorln(e)
	}

	if len(errs) > 0 {
		return ErrWrongSpec
	}

	return nil
}

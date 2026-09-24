package toolversions

import (
	"errors"

	"github.com/sirupsen/logrus"
)

/*
"toolversions" defines the specification for manipulating .tool-versions files.
It can be used as a "source", a "condition", or a "target".
*/
type Spec struct {
	// "file" defines the path of the .tool-versions file to use.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * "file" and "files" are mutually exclusive.
	//   * the schemes "https://" and "http://" are not supported in a target.
	//
	// example:
	//   * file: .tool-versions
	//
	File string `yaml:",omitempty"`
	// "files" defines the list of .tool-versions file paths to use.
	//
	// compatible:
	//   * condition
	//   * target
	//
	// remark:
	//   * "file" and "files" are mutually exclusive.
	//
	Files []string `yaml:",omitempty"`
	// "key" defines the name of the tool to use.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * "key" is required.
	//
	// example:
	//   * key: golang
	//
	Key string `yaml:",omitempty"`
	// "value" defines the version associated with the tool.
	//
	// compatible:
	//   * condition
	//   * target
	//
	// default:
	//   the output of the associated source.
	//
	Value string `yaml:",omitempty"`
	// "createmissingkey" creates the key when the .tool-versions file does not hold it yet.
	//
	// compatible:
	//   * target
	//
	// default:
	//   false
	//
	// remark:
	//   * when false, a missing key raises an error.
	//
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

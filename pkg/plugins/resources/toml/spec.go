package toml

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

type Spec struct {
	// [s][c][t] File specifies the toml file to manipulate
	File string `yaml:",omitempty"`
	// [c][t] Files specifies a list of Json file to manipulate
	Files []string `yaml:",omitempty"`
	// [s][c][t] Query allows to used advanced query. Override the parameter key
	Query string `yaml:",omitempty"`
	// [s][c][t] Key specifies the query to retrieve an information from a toml file
	Key string `yaml:",omitempty"`
	// [s][c][t] Value specifies the value for a specific key. Default to source output
	Value string `yaml:",omitempty"`
	// [c][t] *Deprecated* Please look at query parameter to achieve similar objective
	Multiple bool `yaml:",omitempty" jsonschema:"-"`
	// [s]VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
}

var (
	// ErrSpecFileUndefined is returned if a file wasn't specified
	ErrSpecFileUndefined = errors.New("toml file undefined")
	// ErrSpecKeyUndefined is returned if a key wasn't specified
	ErrSpecKeyUndefined = errors.New("toml key or query undefined")
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
	if len(s.Key) == 0 && len(s.Query) == 0 {
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

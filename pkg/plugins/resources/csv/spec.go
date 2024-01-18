package csv

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

type Spec struct {
	// [s][c][t] File specifies the csv file
	File string `yaml:",omitempty"`
	// [c][t] Files specifies a list of Json file to manipulate
	Files []string `yaml:",omitempty"`
	// [s][c][t] Key specifies the csv query
	Key string `yaml:",omitempty"`
	// [s][c][t] Query allows to used advanced query. Override the parameter key
	Query string `yaml:",omitempty"`
	// [s][c][t] Key specifies the csv value, default to source output
	Value string `yaml:",omitempty"`
	// [s][c][t] Comma specifies the csv separator character, default ","
	Comma rune `yaml:",omitempty"`
	// [s][c][t] Comma specifies the csv comment character, default "#"
	Comment rune `yaml:",omitempty"`
	// [c][t] *Deprecated* Please look at query parameter to achieve similar objective
	Multiple bool `yaml:",omitempty" jsonschema:"-"`
	// [s]VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
}

var (
	ErrSpecFileUndefined       = errors.New("csv file undefined")
	ErrSpecKeyUndefined        = errors.New("csv key or query undefined")
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

	if len(errs) > 0 {
		for i := range errs {
			logrus.Errorln(errs[i])
		}
		return ErrWrongSpec
	}

	return nil
}

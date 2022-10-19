package json

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

type Spec struct {
	// [s][c][t] File specifies the Json file to manipuate
	File string `yaml:",omitempty"`
	// [c][t] Files specifies a list of Json file to manipuate
	Files []string `yaml:",omitempty"`
	// [s][c][t] Key specifies the Jsonpath key to manipuate
	Key string `yaml:",omitempty"`
	// [s][c][t] Value specifies the Jsonpath key to manipuate. Default to source output
	Value string `yaml:",omitempty"`
	// [s][c][t] Multiple allows to query multiple values at once
	Multiple bool `yaml:",omitempty"`
	// [s]VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
}

var (
	ErrSpecFileUndefined                = errors.New("json file undefined")
	ErrSpecKeyUndefined                 = errors.New("json key undefined")
	ErrSpecFileAndFilesDefined          = errors.New("parameter \"file\" and \"files\" are mutually exclusive")
	ErrSpecVersionFilterRequireMultiple = errors.New("parameter \"versionfilter\" and \"multiple\" must be used together")
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

	if (s.Multiple && s.VersionFilter.IsZero()) ||
		(!s.Multiple && !s.VersionFilter.IsZero()) {
		errs = append(errs, ErrSpecVersionFilterRequireMultiple)
	}

	for _, e := range errs {
		logrus.Errorln(e)
	}

	if len(errs) > 0 {
		return ErrWrongSpec
	}

	return nil
}

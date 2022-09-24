package json

import "errors"

type Spec struct {
	File  string `yaml:",omitempty"`
	Key   string `yaml:",omitempty"`
	Value string `yaml:",omitempty"`
}

var (
	ErrSpecFileUndefined = errors.New("json file not specified")
	ErrSpecKeyUndefined  = errors.New("json key undefined")
)

func (s *Spec) Validate() (errs []error) {
	if len(s.File) == 0 {
		errs = append(errs, errors.New(""))
	}
	if len(s.Key) == 0 {
		errs = append(errs, errors.New("json key not specified "))
	}
	return errs
}

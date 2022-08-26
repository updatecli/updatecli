package xml

import "errors"

type Spec struct {
	File  string `yaml:",omitempty"`
	Path  string `yaml:",omitempty"`
	Value string `yaml:",omitempty"`
}

var (
	ErrSpecFileUndefined = errors.New("xml file not specified")
	ErrSpecKeyUndefined  = errors.New("xml key undefined")
)

func (s *Spec) Validate() (errs []error) {
	if len(s.File) == 0 {
		errs = append(errs, errors.New(""))
	}
	if len(s.Path) == 0 {
		errs = append(errs, errors.New("xml key not specified "))
	}
	return errs
}

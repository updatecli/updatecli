package xml

import "errors"

type Spec struct {
	// [s][c][t] File defines the xml path to interact with
	File string `yaml:",omitempty"`
	// [s][c][t] Path defines the xmlPAth used for doing the query
	Path string `yaml:",omitempty"`
	// [s][c][t] Value specifies the value for a specific Path. Default value fetch from source input
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

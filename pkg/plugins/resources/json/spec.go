package json

import "errors"

type Spec struct {
	// [s][c][t] File specifies the Json file to manipuate
	File string `yaml:",omitempty"`
	// [s][c][t] Key specifies the Jsonpath key to manipuate
	Key string `yaml:",omitempty"`
	// [s][c][t] Value specifies the Jsonpath key to manipuate. Default to source output
	Value string `yaml:",omitempty"`
}

var (
	ErrSpecFileUndefined = errors.New("json file not specified")
	ErrSpecKeyUndefined  = errors.New("json key undefined")
)

func (s *Spec) Validate() (errs []error) {
	if len(s.File) == 0 {
		errs = append(errs, ErrSpecFileUndefined)
	}
	if len(s.Key) == 0 {
		errs = append(errs, ErrSpecKeyUndefined)
	}
	return errs
}

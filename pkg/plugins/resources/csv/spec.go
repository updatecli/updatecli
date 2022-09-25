package csv

import "errors"

type Spec struct {
	// [s][c][t] File specifies the csv file
	File string `yaml:",omitempty"`
	// [s][c][t] Key specifies the csv query
	Key string `yaml:",omitempty"`
	// [s][c][t] Key specifies the csv value, default to source output
	Value string `yaml:",omitempty"`
	// [s][c][t] Comma specifies the csv separator character, default ","
	Comma rune `yaml:",omitempty"`
	// [s][c][t] Comma specifies the csv comment character, default "#"
	Comment rune `yaml:",omitempty"`
}

var (
	ErrSpecFileUndefined = errors.New("csv file not specified")
	ErrSpecKeyUndefined  = errors.New("csv key undefined")
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

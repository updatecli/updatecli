package toml

import "errors"

type Spec struct {
	// [s][c][t] File specifies the toml file to manipulate
	File string `yaml:",omitempty"`
	// [s][c][t] Key specifies the query to retrieve an information from a toml file
	Key string `yaml:",omitempty"`
	// [s][c][t] Value specifies the value for a specific key. Default to source output
	Value string `yaml:",omitempty"`
	// [c][t] Multiple allows to query multiple values at once
	Multiple bool `yaml:",omitempty"`
}

var (
	// ErrSpecFileUndefined is returned if a file wasn't specified
	ErrSpecFileUndefined = errors.New("toml file not specified")
	// ErrSpecKeyUndefined is returned if a key wasn't specified
	ErrSpecKeyUndefined = errors.New("toml key undefined")
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

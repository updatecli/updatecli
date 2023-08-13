package xml

import "errors"

/*
"xml" defines the specification for manipulating "xml" files.
It can be used as a "source", a "condition", or a "target".
*/
type Spec struct {
	/*
		"file" define the xml file path to interact with.

		compatible:
			* source
			* condition
			* target

		remark:
			* scheme "https://", "http://", and "file://" are supported in path for source and condition
	*/
	File string `yaml:",omitempty"`
	/*
		"path" defines the xml path used for doing the query

		compatible:
			* source
			* condition
			* target

		example:
			* path: "/project/parent/version"
			* path: "//breakfast_menu/food[0]/name"
			* path: "//book[@category='WEB']/title"
	*/
	Path string `yaml:",omitempty"`
	/*
		"value" is the value associated with a xmlpath query.

		compatible:
			* source
			* condition
			* target

	*/
	Value string `yaml:",omitempty"`
}

var (
	// ErrSpecFileUndefined is returned when the file is not defined
	ErrSpecFileUndefined = errors.New("xml file not specified")
	// ErrSpecPathUndefined is returned when the path is not defined
	ErrSpecPathUndefined = errors.New("xml path undefined")
)

func (s *Spec) Validate() (errs []error) {
	if len(s.File) == 0 {
		errs = append(errs, errors.New(""))
	}
	if len(s.Path) == 0 {
		errs = append(errs, ErrSpecPathUndefined)
	}
	return errs
}

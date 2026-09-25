package xml

import "errors"

/*
"xml" defines the specification for manipulating xml files.
It can be used as a "source", a "condition", or a "target".
*/
type Spec struct {
	// "file" defines the path of the xml file to use.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * "file" is required.
	//   * the schemes "https://", "http://" and "file://" are supported in a source or a condition.
	//
	// example:
	//   * file: pom.xml
	//
	File string `yaml:",omitempty"`
	// "path" defines the xpath query used to retrieve a value from the xml document.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * "path" is required.
	//
	// example:
	//   * path: "/project/parent/version"
	//   * path: "//breakfast_menu/food[0]/name"
	//   * path: "//book[@category='WEB']/title"
	//
	Path string `yaml:",omitempty"`
	// "value" defines the value associated with the xpath query.
	//
	// compatible:
	//   * condition
	//   * target
	//
	// default:
	//   the output of the associated source.
	//
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
		errs = append(errs, ErrSpecFileUndefined)
	}
	if len(s.Path) == 0 {
		errs = append(errs, ErrSpecPathUndefined)
	}
	return errs
}

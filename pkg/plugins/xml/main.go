package xml

import (
	"errors"

	"github.com/sirupsen/logrus"
)

var (
	// ErrDaselFailedParsingXMLByteFormat is returned if dasel couldn't parse the byteData
	ErrDaselFailedParsingXMLByteFormat error = errors.New("fail to parse XML data")
	// ErrWrongSpec is returned when the Spec has wrong content
	ErrWrongSpec error = errors.New("wrong spec content")
)

// XML stores configuration about the file and the key value which needs to be updated.
type XML struct {
	spec Spec
}

func New(s *Spec) (*XML, error) {
	x := XML{}

	x.spec = *s

	err := x.Validate()

	return &x, err
}

func (x *XML) Validate() (err error) {
	errs := x.spec.Validate()

	if len(errs) > 0 {
		for _, e := range errs {
			logrus.Errorf(e.Error())
		}
		return ErrWrongSpec
	}
	return nil
}

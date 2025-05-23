package xml

import (
	"errors"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/text"
)

var (
	// ErrWrongSpec is returned when the Spec has wrong content
	ErrWrongSpec error = errors.New("wrong spec content")
)

// XML stores configuration about the file and the key value which needs to be updated.
type XML struct {
	spec             Spec
	contentRetriever text.TextRetriever
	currentContent   string
}

func New(spec interface{}) (*XML, error) {

	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	newSpec.File = strings.TrimPrefix(newSpec.File, "file://")

	x := XML{
		spec:             newSpec,
		contentRetriever: &text.Text{},
	}

	err = x.Validate()

	if err != nil {
		return nil, err
	}

	return &x, err
}

// Read reads the file content
func (x *XML) Read(filename string) error {
	textContent, err := x.contentRetriever.ReadAll(filename)
	if err != nil {
		return err
	}
	x.currentContent = textContent
	return nil
}

// Validate checks if the Spec is correctly defined
func (x *XML) Validate() (err error) {
	errs := x.spec.Validate()

	if len(errs) > 0 {
		for _, e := range errs {
			logrus.Errorln(e.Error())
		}
		return ErrWrongSpec
	}
	return nil
}

// ReportConfig returns a new configuration object with only the necessary fields
// to identify the resource without any sensitive information or context specific data.
func (x *XML) ReportConfig() interface{} {
	return Spec{
		File:  x.spec.File,
		Path:  x.spec.Path,
		Value: x.spec.Value,
	}
}

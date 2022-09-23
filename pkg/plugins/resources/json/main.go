package json

import (
	"errors"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/text"
)

var (
	// ErrDaselFailedParsingXMLByteFormat is returned if dasel couldn't parse the byteData
	ErrDaselFailedParsingJSONByteFormat error = errors.New("fail to parse Json data")
	// ErrWrongSpec is returned when the Spec has wrong content
	ErrWrongSpec error = errors.New("wrong spec content")
)

// Json stores configuration about the file and the key value which needs to be updated.
type Json struct {
	spec             Spec
	contentRetriever text.TextRetriever
	currentContent   string
}

func New(spec interface{}) (*Json, error) {

	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	newSpec.File = strings.TrimPrefix(newSpec.File, "file://")

	j := Json{
		spec:             newSpec,
		contentRetriever: &text.Text{},
	}

	err = j.Validate()

	if err != nil {
		return nil, err
	}

	return &j, err
}

func (j *Json) Validate() (err error) {
	errs := j.spec.Validate()

	if len(errs) > 0 {
		for _, e := range errs {
			logrus.Errorf(e.Error())
		}
		return ErrWrongSpec
	}
	return nil
}

func (j *Json) Read() error {
	textContent, err := j.contentRetriever.ReadAll(j.spec.File)
	if err != nil {
		return err
	}
	j.currentContent = textContent
	return nil
}

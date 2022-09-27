package toml

import (
	"errors"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/text"
)

var (
	// ErrDaselFailedParsingTOMLByteFormat is returned if dasel couldn't parse the byteData
	ErrDaselFailedParsingTOMLByteFormat error = errors.New("fail to parse Toml data")
	// ErrWrongSpec is returned when the Spec has wrong content
	ErrWrongSpec error = errors.New("wrong spec content")
)

// Toml stores configuration about the file and the key value which needs to be updated.
type Toml struct {
	spec             Spec
	contentRetriever text.TextRetriever
	currentContent   string
}

func New(spec interface{}) (*Toml, error) {

	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	newSpec.File = strings.TrimPrefix(newSpec.File, "file://")

	j := Toml{
		spec:             newSpec,
		contentRetriever: &text.Text{},
	}

	err = j.Validate()

	if err != nil {
		return nil, err
	}

	return &j, err
}

func (t *Toml) Validate() (err error) {
	errs := t.spec.Validate()

	if len(errs) > 0 {
		for _, e := range errs {
			logrus.Errorf(e.Error())
		}
		return ErrWrongSpec
	}
	return nil
}

func (t *Toml) Read() error {
	textContent, err := t.contentRetriever.ReadAll(t.spec.File)
	if err != nil {
		return err
	}
	t.currentContent = textContent
	return nil
}

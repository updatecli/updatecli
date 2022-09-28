package csv

import (
	"errors"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/tomwright/dasel/storage"
	"github.com/updatecli/updatecli/pkg/core/text"
)

var (
	// ErrDaselFailedParsingJSONByteFormat is returned if dasel couldn't parse the byteData
	ErrDaselFailedParsingJSONByteFormat error = errors.New("fail to parse Json data")
	// ErrWrongSpec is returned when the Spec has wrong content
	ErrWrongSpec error = errors.New("wrong spec content")
)

const (
	// DEFAULTSEPARATOR defines the default csv separator
	DEFAULTSEPARATOR rune = ','
	// DEFAULTCOMMENT defines the default comment character
	DEFAULTCOMMENT rune = '#'
)

// CSV stores configuration about the file and the key value which needs to be updated.
type CSV struct {
	spec             Spec
	contentRetriever text.TextRetriever
	currentContent   string
	csvDocument      storage.CSVDocument
}

func New(spec interface{}) (*CSV, error) {

	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	var emptyRune rune
	if newSpec.Comma == emptyRune {
		newSpec.Comma = DEFAULTSEPARATOR
	}

	if newSpec.Comment == emptyRune {
		newSpec.Comment = DEFAULTCOMMENT
	}

	newSpec.File = strings.TrimPrefix(newSpec.File, "file://")

	c := CSV{
		spec:             newSpec,
		contentRetriever: &text.Text{},
	}

	err = c.Validate()

	if err != nil {
		return nil, err
	}

	return &c, err
}

func (c *CSV) Validate() (err error) {
	errs := c.spec.Validate()

	if len(errs) > 0 {
		for _, e := range errs {
			logrus.Errorf(e.Error())
		}
		return ErrWrongSpec
	}
	return nil
}

func (c *CSV) Read() error {
	textContent, err := c.contentRetriever.ReadAll(c.spec.File)
	if err != nil {
		return err
	}
	c.currentContent = textContent
	return nil
}

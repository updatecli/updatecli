package csv

import (
	"errors"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/utils/dasel"
)

var (
	// ErrDaselFailedParsingJSONByteFormat is returned if dasel couldn't parse the byteData
	ErrDaselFailedParsingJSONByteFormat error = errors.New("fail to parse Json data")
)

const (
	// DEFAULTSEPARATOR defines the default csv separator
	DEFAULTSEPARATOR rune = ','
	// DEFAULTCOMMENT defines the default comment character
	DEFAULTCOMMENT rune = '#'
)

// CSV stores configuration about the file and the key value which needs to be updated.
type CSV struct {
	spec Spec
	//contentRetriever text.TextRetriever
	//currentContent   string
	//csvDocument      storage.CSVDocument
	contents []csvContent
}

func New(spec interface{}) (*CSV, error) {

	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	comma := DEFAULTSEPARATOR
	comment := DEFAULTCOMMENT
	var emptyRune rune
	if newSpec.Comma != emptyRune {
		comma = newSpec.Comma
	}

	if newSpec.Comment != emptyRune {
		comment = newSpec.Comment
	}

	newSpec.File = strings.TrimPrefix(newSpec.File, "file://")

	if err := newSpec.Validate(); err != nil {
		return nil, err
	}

	c := CSV{
		spec: newSpec,
	}

	// Init currentContents
	switch len(c.spec.File) > 0 {
	case true:
		c.contents = append(
			c.contents, csvContent{
				comma:   comma,
				comment: comment,
				FileContent: dasel.FileContent{
					DataType:         "csv",
					FilePath:         c.spec.File,
					ContentRetriever: &text.Text{},
				},
			})

	case false:
		for i := range c.spec.Files {
			c.contents = append(
				c.contents, csvContent{
					comma:   comma,
					comment: comment,
					FileContent: dasel.FileContent{
						DataType:         "csv",
						FilePath:         c.spec.Files[i],
						ContentRetriever: &text.Text{},
					},
				})
		}
	}

	return &c, err
}

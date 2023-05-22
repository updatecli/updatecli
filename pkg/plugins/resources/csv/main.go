package csv

import (
	"errors"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/utils/dasel"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
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
	spec     Spec
	contents []csvContent
	// Holds both parsed version and original version (to allow retrieving metadata such as changelog)
	foundVersion version.Version
	// Holds the "valid" version.filter, that might be different than the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
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

	// Deprecate message to remove in a future updatecli version
	// cfr https://github.com/updatecli/updatecli/pull/944
	if newSpec.Multiple {
		logrus.Warningln("the setting 'multiple' is now deprecated. you should use the parameter \"query\" which allows you to specify an advanced query.")
		if len(newSpec.Key) > 0 && len(newSpec.Query) == 0 {
			logrus.Printf("Instead of using the parameter key %q combined with multiple, we converted your setting to use the parameter query %q", newSpec.Key, newSpec.Query)
			newSpec.Query = newSpec.Key
			newSpec.Key = ""
			newSpec.Multiple = false
		}
	}

	newFilter, err := newSpec.VersionFilter.Init()
	if err != nil {
		return nil, err
	}

	c := CSV{
		spec:          newSpec,
		versionFilter: newFilter,
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

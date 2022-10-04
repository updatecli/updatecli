package toml

import (
	"errors"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/utils/dasel"
)

var (
	// ErrDaselFailedParsingTOMLByteFormat is returned if dasel couldn't parse the byteData
	ErrDaselFailedParsingTOMLByteFormat error = errors.New("fail to parse Toml data")
)

// Toml stores configuration about the file and the key value which needs to be updated.
type Toml struct {
	spec     Spec
	contents []dasel.FileContent
}

func New(spec interface{}) (*Toml, error) {

	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	err = newSpec.Validate()

	newSpec.File = strings.TrimPrefix(newSpec.File, "file://")

	j := Toml{
		spec: newSpec,
	}

	// Init currentContents
	switch len(j.spec.File) > 0 {
	case true:
		j.contents = append(
			j.contents, dasel.FileContent{
				DataType:         "toml",
				FilePath:         j.spec.File,
				ContentRetriever: &text.Text{},
			})

	case false:
		for i := range j.spec.Files {
			j.contents = append(
				j.contents, dasel.FileContent{
					DataType:         "toml",
					FilePath:         j.spec.Files[i],
					ContentRetriever: &text.Text{},
				})
		}
	}

	return &j, err
}

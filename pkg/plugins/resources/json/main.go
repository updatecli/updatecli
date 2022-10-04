package json

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/utils/dasel"
)

// Json stores configuration about the file and the key value which needs to be updated.
type Json struct {
	spec     Spec
	contents []dasel.FileContent
}

func New(spec interface{}) (*Json, error) {

	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	err = newSpec.Validate()

	newSpec.File = strings.TrimPrefix(newSpec.File, "file://")

	j := Json{
		spec: newSpec,
	}

	// Init currentContents
	switch len(j.spec.File) > 0 {
	case true:
		j.contents = append(
			j.contents, dasel.FileContent{
				DataType:         "json",
				FilePath:         j.spec.File,
				ContentRetriever: &text.Text{},
			})

	case false:
		for i := range j.spec.Files {
			j.contents = append(
				j.contents, dasel.FileContent{
					DataType:         "json",
					FilePath:         j.spec.Files[i],
					ContentRetriever: &text.Text{},
				})
		}
	}

	if err != nil {
		return nil, err
	}

	return &j, err
}

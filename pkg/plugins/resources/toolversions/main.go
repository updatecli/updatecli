package toolversions

import (
	"errors"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/utils/toolversions"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

var (
	// ErrDaselFailedParsingToolsVersionsByteFormat is returned if dasel couldn't parse the byteData
	ErrDaselFailedParsingToolsVersionsByteFormat error = errors.New("fail to parse .tool-versions data")
)

// ToolVersions stores configuration about the file and the key value which needs to be updated.
type ToolVersions struct {
	spec     Spec
	contents []toolversions.FileContent
	// Holds both parsed version and original version (to allow retrieving metadata such as changelog)
	foundVersion version.Version
}

func New(spec interface{}) (*ToolVersions, error) {

	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	err = newSpec.Validate()

	if err != nil {
		return nil, err
	}

	newSpec.File = strings.TrimPrefix(newSpec.File, "file://")

	j := ToolVersions{
		spec: newSpec,
	}

	// Init currentContents
	switch len(j.spec.File) > 0 {
	case true:
		j.contents = append(
			j.contents, toolversions.FileContent{
				FilePath:         j.spec.File,
				ContentRetriever: &text.Text{},
			})

	case false:
		for i := range j.spec.Files {
			j.contents = append(
				j.contents, toolversions.FileContent{
					FilePath:         j.spec.Files[i],
					ContentRetriever: &text.Text{},
				})
		}
	}

	return &j, err
}

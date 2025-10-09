package toolversions

import (
	"errors"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/utils/toolversions"
)

var (
	// ErrDaselFailedParsingToolsVersionsByteFormat is returned if dasel couldn't parse the byteData
	ErrDaselFailedParsingToolsVersionsByteFormat error = errors.New("fail to parse .tool-versions data")
)

// ToolVersions stores configuration about the file and the key value which needs to be updated.
type ToolVersions struct {
	spec     Spec
	contents []toolversions.FileContent
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

// ReportConfig returns a new configuration object with only the necessary fields
// to identify the resource without any sensitive information or context specific data.
func (s *ToolVersions) ReportConfig() interface{} {
	return Spec{
		File:             s.spec.File,
		Files:            s.spec.Files,
		Key:              s.spec.Key,
		Value:            s.spec.Value,
		CreateMissingKey: s.spec.CreateMissingKey,
	}
}

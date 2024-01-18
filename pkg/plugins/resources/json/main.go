package json

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/utils/dasel"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Json stores configuration about the file and the key value which needs to be updated.
type Json struct {
	spec     Spec
	contents []dasel.FileContent
	// Holds both parsed version and original version (to allow retrieving metadata such as changelog)
	foundVersion version.Version
	// Holds the "valid" version.filter, that might be different than the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
}

func New(spec interface{}) (*Json, error) {

	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	err = newSpec.Validate()

	if err != nil {
		return nil, err
	}

	// Deprecate message to remove in a future updatecli version
	// cfr https://github.com/updatecli/updatecli/pull/944
	if newSpec.Multiple {
		logrus.Warningln("the setting multiple is now deprecated. you should be using the parameter \"query\" which allows you to specify an advanced query.")
		if len(newSpec.Key) > 0 && len(newSpec.Query) == 0 {
			logrus.Printf("Instead of using the parameter key %q combined with multiple, we converted your setting to use the parameter query %q", newSpec.Key, newSpec.Query)
			newSpec.Query = newSpec.Key
			newSpec.Key = ""
			newSpec.Multiple = false
		}
	}

	newSpec.File = strings.TrimPrefix(newSpec.File, "file://")

	newFilter, err := newSpec.VersionFilter.Init()
	if err != nil {
		return nil, err
	}

	j := Json{
		spec:          newSpec,
		versionFilter: newFilter,
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

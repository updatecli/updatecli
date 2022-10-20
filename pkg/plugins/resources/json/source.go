package json

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

var (
	ErrSpecVersionFilterRequireMultiple = errors.New("in the context of a source, parameter \"versionfilter\" and \"multiple\" must be used together")
)

func (j *Json) Source(workingDir string) (string, error) {

	if len(j.contents) > 1 {
		return "", errors.New("source only supports one file")
	}

	if (j.spec.Multiple && j.spec.VersionFilter.IsZero()) ||
		(!j.spec.Multiple && !j.spec.VersionFilter.IsZero()) {
		return "", ErrSpecVersionFilterRequireMultiple
	}

	content := j.contents[0]

	sourceOutput := ""
	if err := content.Read(workingDir); err != nil {
		return "", err
	}

	switch j.spec.Multiple {
	case true:
		queryResults, err := content.MultipleQuery(j.spec.Key)

		if err != nil {
			return "", err
		}

		j.foundVersion, err = j.versionFilter.Search(queryResults)
		if err != nil {
			return "", err
		}
		sourceOutput = j.foundVersion.GetVersion()

	case false:
		queryResult, err := content.DaselNode.Query(j.spec.Key)
		if err != nil {
			// Catch error message returned by Dasel, if it couldn't find the node
			// This is approach is not very robust
			// https://github.com/TomWright/dasel/blob/master/node_query.go#L58

			if strings.HasPrefix(err.Error(), "could not find value:") {
				err := fmt.Errorf("%s cannot find value for path %q from file %q",
					result.FAILURE,
					j.spec.Key,
					content.FilePath)
				return "", err
			}
			return "", err
		}

		sourceOutput = queryResult.String()
	}

	logrus.Infof("%s Value %q, found in file %q, for key %q'",
		result.SUCCESS,
		sourceOutput,
		content.FilePath,
		j.spec.Key)

	return sourceOutput, nil
}

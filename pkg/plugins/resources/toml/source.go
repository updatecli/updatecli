package toml

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

var (
	ErrSpecVersionFilterRequireMultiple = errors.New("in the context of a source, parameter \"versionfilter\" and \"query\" must be used together")
)

func (t *Toml) Source(workingDir string) (string, error) {

	if len(t.contents) > 1 {
		return "", errors.New("source only supports one file")
	}

	if (len(t.spec.Query) > 0 && t.spec.VersionFilter.IsZero()) ||
		(len(t.spec.Query) == 0) && !t.spec.VersionFilter.IsZero() {
		return "", ErrSpecVersionFilterRequireMultiple
	}

	content := t.contents[0]

	sourceOutput := ""

	if err := content.Read(workingDir); err != nil {
		return "", err
	}

	query := ""
	switch len(t.spec.Query) > 0 {
	case true:
		query = t.spec.Query
		queryResults, err := content.MultipleQuery(query)

		if err != nil {
			return "", err
		}

		t.foundVersion, err = t.versionFilter.Search(queryResults)
		if err != nil {
			return "", err
		}
		sourceOutput = t.foundVersion.GetVersion()

	case false:
		query = t.spec.Key
		queryResult, err := content.DaselNode.Query(query)
		if err != nil {
			// Catch error message returned by Dasel, if it couldn't find the node
			// This is approach is not very robust
			// https://github.com/TomWright/dasel/blob/master/node_query.go#L58

			if strings.HasPrefix(err.Error(), "could not find value:") {
				err := fmt.Errorf("%s cannot find value for path %q from file %q",
					result.FAILURE,
					t.spec.Key,
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
		query)

	return sourceOutput, nil
}

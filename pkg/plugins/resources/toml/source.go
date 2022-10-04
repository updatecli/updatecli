package toml

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (t *Toml) Source(workingDir string) (string, error) {

	if len(t.contents) > 1 {
		return "", errors.New("source only supports one file")
	}

	sourceOutput := ""
	for i := range t.contents {
		if err := t.contents[i].Read(workingDir); err != nil {
			return "", err
		}

		queryResult, err := t.contents[i].DaselNode.Query(t.spec.Key)
		if err != nil {
			// Catch error message returned by Dasel, if it couldn't find the node
			// This is approach is not very robust
			// https://github.com/TomWright/dasel/blob/master/node_query.go#L58

			if strings.HasPrefix(err.Error(), "could not find value:") {
				err := fmt.Errorf("%s cannot find value for path %q from file %q",
					result.FAILURE,
					t.spec.Key,
					t.contents[i].FilePath)
				return "", err
			}
			return "", err
		}

		logrus.Infof("%s Value %q, found in file %q, for key %q'",
			result.SUCCESS,
			queryResult.String(),
			t.contents[i].FilePath,
			t.spec.Key)

		sourceOutput = queryResult.String()

	}

	return sourceOutput, nil
}

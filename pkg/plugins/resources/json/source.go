package json

import (
	"errors"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (j *Json) Source(workingDir string) (string, error) {

	if len(j.contents) > 1 {
		return "", errors.New("source only supports one file")
	}

	sourceOutput := ""
	for i := range j.contents {
		if err := j.contents[i].Read(workingDir); err != nil {
			return "", err
		}

		queryResult, err := j.contents[i].DaselNode.Query(j.spec.Key)
		if err != nil {
			// Catch error message returned by Dasel, if it couldn't find the node
			// This is approach is not very robust
			// https://github.com/TomWright/dasel/blob/master/node_query.go#L58

			if strings.HasPrefix(err.Error(), "could not find value:") {
				logrus.Infof("%s cannot find value for path %q from file %q",
					result.FAILURE,
					j.spec.Key,
					j.contents[i].FilePath)
				return "", nil
			}
			return "", err
		}

		logrus.Infof("%s Value %q, found in file %q, for key %q'",
			result.SUCCESS,
			queryResult.String(),
			j.contents[i].FilePath,
			j.spec.Key)

		sourceOutput = queryResult.String()

	}

	return sourceOutput, nil
}

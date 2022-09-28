package json

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/tomwright/dasel"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (j *Json) Source(workingDir string) (string, error) {

	// Test at runtime if a file exist
	if !j.contentRetriever.FileExists(j.spec.File) {
		return "", fmt.Errorf("the Json file %q does not exist", j.spec.File)
	}

	if err := j.Read(); err != nil {
		return "", err
	}

	var data interface{}

	err := json.Unmarshal([]byte(j.currentContent), &data)

	if err != nil {
		return "", err
	}

	rootNode := dasel.New(data)

	if rootNode == nil {
		return "", ErrDaselFailedParsingJSONByteFormat
	}

	queryResult, err := rootNode.Query(j.spec.Key)
	if err != nil {
		// Catch error message returned by Dasel, if it couldn't find the node
		// This is approach is not very robust
		// https://github.com/TomWright/dasel/blob/master/node_query.go#L58

		if strings.HasPrefix(err.Error(), "could not find value:") {
			logrus.Infof("%s cannot find value for path %q from file %q",
				result.FAILURE,
				j.spec.Key,
				j.spec.File)
			return "", nil
		}
		return "", err
	}

	logrus.Infof("%s Value %q, found in file %q, for key %q'",
		result.SUCCESS,
		queryResult.String(),
		j.spec.File,
		j.spec.Key)

	return queryResult.String(), nil
}

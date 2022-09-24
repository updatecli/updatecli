package json

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/tomwright/dasel"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (j *Json) Condition(source string) (bool, error) {
	return j.ConditionFromSCM(source, nil)
}

func (j *Json) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	if scm != nil {
		j.spec.File = joinPathWithWorkingDirectoryPath(j.spec.File, scm.GetDirectory())
	}

	// Test at runtime if a file exist
	if !j.contentRetriever.FileExists(j.spec.File) {
		return false, fmt.Errorf("the Json file %q does not exist", j.spec.File)
	}

	if err := j.Read(); err != nil {
		return false, err
	}

	// Override value from source if not yet defined
	if len(j.spec.Value) == 0 {
		j.spec.Value = source
	}

	var data interface{}

	err := json.Unmarshal([]byte(j.currentContent), &data)

	if err != nil {
		return false, err
	}

	rootNode := dasel.New(data)

	if rootNode == nil {
		return false, ErrDaselFailedParsingJSONByteFormat
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
			return false, nil
		}

		return false, err
	}

	if queryResult.String() == j.spec.Value {
		logrus.Infof("%s Key '%s', from file '%v', is correctly set to %s'",
			result.SUCCESS,
			j.spec.Key,
			j.spec.File,
			j.spec.Value)
		return true, nil
	} else {
		logrus.Infof("%s Key '%s', from file '%v', is incorrectly set to %s and should be %s'",
			result.ATTENTION,
			j.spec.Key,
			j.spec.File,
			queryResult.String(),
			j.spec.Value)
	}
	return false, nil
}

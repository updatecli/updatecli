package json

import (
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (j *Json) Condition(source string) (bool, error) {
	return j.ConditionFromSCM(source, nil)
}

func (j *Json) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	conditionResult := true

	rootDir := ""
	if scm != nil {
		rootDir = scm.GetDirectory()
	}

	for i := range j.contents {

		if err := j.contents[i].Read(rootDir); err != nil {
			return false, err
		}

		// Override value from source if not yet defined
		if len(j.spec.Value) == 0 {
			j.spec.Value = source
		}

		var queryResults []string
		var err error

		switch len(j.spec.Query) > 0 {
		case true:
			queryResults, err = j.contents[i].MultipleQuery(j.spec.Query)

			if err != nil {
				return false, err
			}

		case false:
			queryResult, err := j.contents[i].Query(j.spec.Key)

			if err != nil {
				return false, err
			}

			queryResults = []string{queryResult}
		}

		for _, queryResult := range queryResults {
			switch queryResult == j.spec.Value {
			case true:
				logrus.Infof("%s Key %q, from file %q, is correctly set to %q'",
					result.SUCCESS,
					j.spec.Key,
					j.contents[i].FilePath,
					j.spec.Value)

			case false:
				conditionResult = false
				logrus.Infof("%s Key %q, from file %q, is incorrectly set to %q and should be %q",
					result.ATTENTION,
					j.spec.Key,
					j.contents[i].FilePath,
					queryResult,
					j.spec.Value)
			}
		}

	}
	return conditionResult, nil
}

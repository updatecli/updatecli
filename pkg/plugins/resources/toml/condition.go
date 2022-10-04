package toml

import (
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (t *Toml) Condition(source string) (bool, error) {
	return t.ConditionFromSCM(source, nil)
}

func (t *Toml) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	conditionResult := true

	rootDir := ""
	if scm != nil {
		rootDir = scm.GetDirectory()
	}

	for i := range t.contents {

		if err := t.contents[i].Read(rootDir); err != nil {
			return false, err
		}

		// Override value from source if not yet defined
		if len(t.spec.Value) == 0 {
			t.spec.Value = source
		}

		var queryResults []string
		var err error

		switch t.spec.Multiple {
		case true:
			queryResults, err = t.contents[i].MultipleQuery(t.spec.Key)

			if err != nil {
				return false, err
			}

		case false:
			queryResult, err := t.contents[i].Query(t.spec.Key)

			if err != nil {
				return false, err
			}

			queryResults = []string{queryResult}
		}

		for _, queryResult := range queryResults {
			switch queryResult == t.spec.Value {
			case true:
				logrus.Infof("%s Key %q, from file %q, is correctly set to %q'",
					result.SUCCESS,
					t.spec.Key,
					t.contents[i].FilePath,
					t.spec.Value)

			case false:
				conditionResult = false
				logrus.Infof("%s Key %q, from file %q, is incorrectly set to %q and should be %q",
					result.ATTENTION,
					t.spec.Key,
					t.contents[i].FilePath,
					queryResult,
					t.spec.Value)
			}
		}

	}
	return conditionResult, nil
}

package toml

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (t *Toml) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {
	conditionResult := true

	rootDir := ""
	if scm != nil {
		rootDir = scm.GetDirectory()
	}

	for i := range t.contents {

		if err := t.contents[i].Read(rootDir); err != nil {
			return fmt.Errorf("reading toml file: %w", err)
		}

		// Override value from source if not yet defined
		if len(t.spec.Value) == 0 {
			t.spec.Value = source
		}

		var queryResults []string
		var err error

		switch len(t.spec.Query) > 0 {
		case true:
			queryResults, err = t.contents[i].MultipleQuery(t.spec.Query)

			if err != nil {
				return err
			}

		case false:
			queryResult, err := t.contents[i].Query(t.spec.Key)

			if err != nil {
				return err
			}

			queryResults = []string{queryResult}
		}

		for _, queryResult := range queryResults {
			switch queryResult == t.spec.Value {
			case true:
				logrus.Infof("%s\nkey %q, from file %q, is correctly set to %q",
					resultCondition.Description,
					t.spec.Key,
					t.contents[i].FilePath,
					t.spec.Value)

			case false:
				conditionResult = false
				logrus.Infof("%s\nkey %q, from file %q, is incorrectly set to %q and should be %q",
					resultCondition.Description,
					t.spec.Key,
					t.contents[i].FilePath,
					queryResult,
					t.spec.Value)
			}
		}

	}

	resultCondition.Description = strings.TrimPrefix(resultCondition.Description, "\n")

	if conditionResult {
		resultCondition.Pass = true
		resultCondition.Result = result.SUCCESS
		return nil
	}

	resultCondition.Pass = false
	resultCondition.Result = result.FAILURE

	return nil
}

package csv

import (
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (c *CSV) Condition(source string) (bool, error) {
	return c.ConditionFromSCM(source, nil)
}

func (c *CSV) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	conditionResult := true

	rootDir := ""
	if scm != nil {
		rootDir = scm.GetDirectory()
	}

	for i := range c.contents {

		if err := c.contents[i].Read(rootDir); err != nil {
			return false, err
		}

		// Override value from source if not yet defined
		if len(c.spec.Value) == 0 {
			c.spec.Value = source
		}

		var queryResults []string
		var err error

		switch c.spec.Multiple {
		case true:
			queryResults, err = c.contents[i].MultipleQuery(c.spec.Key)

			if err != nil {
				return false, err
			}

		case false:
			queryResult, err := c.contents[i].Query(c.spec.Key)

			if err != nil {
				return false, err
			}

			queryResults = []string{queryResult}
		}

		for _, queryResult := range queryResults {
			switch queryResult == c.spec.Value {
			case true:
				logrus.Infof("%s Key %q, from file %q, is correctly set to %q'",
					result.SUCCESS,
					c.spec.Key,
					c.contents[i].FilePath,
					c.spec.Value)

			case false:
				conditionResult = false
				logrus.Infof("%s Key %q, from file %q, is incorrectly set to %q and should be %q",
					result.ATTENTION,
					c.spec.Key,
					c.contents[i].FilePath,
					queryResult,
					c.spec.Value)
			}
		}

	}
	return conditionResult, nil

}

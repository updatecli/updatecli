package csv

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (c *CSV) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {

	rootDir := ""
	if scm != nil {
		rootDir = scm.GetDirectory()
	}

	conditionResult := true

	for i := range c.contents {

		if err := c.contents[i].Read(rootDir); err != nil {
			return fmt.Errorf("reading csv file: %w", err)
		}

		// Override value from source if not yet defined
		if len(c.spec.Value) == 0 {
			c.spec.Value = source
		}

		var queryResults []string
		var err error

		switch len(c.spec.Query) > 0 {
		case true:
			queryResults, err = c.contents[i].MultipleQuery(c.spec.Query)
			if err != nil {
				return fmt.Errorf("running queries: %w", err)
			}

		case false:
			queryResult, err := c.contents[i].Query(c.spec.Key)
			if err != nil {
				return fmt.Errorf("running query: %w", err)
			}

			queryResults = []string{queryResult}
		}

		for _, queryResult := range queryResults {
			switch queryResult == c.spec.Value {
			case true:
				resultCondition.Description = fmt.Sprintf("%s\nkey %q, from file %q, is correctly set to %q",
					resultCondition.Description,
					c.spec.Key,
					c.contents[i].FilePath,
					c.spec.Value)

			case false:
				conditionResult = false
				resultCondition.Description = fmt.Sprintf("%s\nkey %q, from file %q, is incorrectly set to %q and should be %q",
					resultCondition.Description,
					c.spec.Key,
					c.contents[i].FilePath,
					queryResult,
					c.spec.Value)
			}
		}
	}

	switch conditionResult {
	case true:
		resultCondition.Pass = true
		resultCondition.Result = result.SUCCESS
	case false:
		resultCondition.Pass = false
		resultCondition.Result = result.FAILURE
	}

	return nil
}

package csv

import (
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func (c *CSV) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {

	rootDir := ""
	if scm != nil {
		rootDir = scm.GetDirectory()
	}

	conditionResult := true
	messages := []string{}

	for i := range c.contents {

		if err := c.contents[i].Read(rootDir); err != nil {
			return false, "", fmt.Errorf("reading csv file: %w", err)
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
				return false, "", fmt.Errorf("running queries: %w", err)
			}

		case false:
			queryResult, err := c.contents[i].Query(c.spec.Key)
			if err != nil {
				return false, "", fmt.Errorf("running query: %w", err)
			}

			queryResults = []string{queryResult}
		}

		for _, queryResult := range queryResults {
			switch queryResult == c.spec.Value {
			case true:
				messages = append(messages, fmt.Sprintf("\nkey %q, from file %q, is correctly set to %q",
					c.spec.Key,
					c.contents[i].FilePath,
					c.spec.Value))

			case false:
				conditionResult = false
				messages = append(messages, fmt.Sprintf("\nkey %q, from file %q, is incorrectly set to %q and should be %q",
					c.spec.Key,
					c.contents[i].FilePath,
					queryResult,
					c.spec.Value))
			}
		}
	}

	return conditionResult, strings.Join(messages, "\n"), nil
}

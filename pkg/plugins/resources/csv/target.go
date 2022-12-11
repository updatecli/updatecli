package csv

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/result"
)

func (c *CSV) Target(source string, dryRun bool) (changed bool, err error) {
	rootDir := ""
	for i := range c.contents {
		filename := c.contents[i].FilePath

		// Target doesn't support updating files on remote http location
		if strings.HasPrefix(filename, "https://") ||
			strings.HasPrefix(filename, "http://") {
			return false, fmt.Errorf("URL scheme is not supported for CSV target: %q", c.spec.File)
		}

		if err := c.contents[i].Read(rootDir); err != nil {
			return false, fmt.Errorf("file %q does not exist", c.contents[i].FilePath)
		}

		if len(c.spec.Value) == 0 {
			c.spec.Value = source
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
				return false, err
			}

		case false:
			queryResult, err := c.contents[i].Query(c.spec.Key)

			if err != nil {
				return false, err
			}

			queryResults = append(queryResults, queryResult)

		}

		for _, queryResult := range queryResults {
			switch queryResult == c.spec.Value {
			case true:
				logrus.Infof("%s Key '%s', from file '%v', is correctly set to %s'",
					result.SUCCESS,
					c.spec.Key,
					c.contents[i].FilePath,
					c.spec.Value)

			case false:
				changed = true
				logrus.Infof("%s Key '%s', from file '%v', is incorrectly set to %s and should be %s'",
					result.ATTENTION,
					c.spec.Key,
					c.contents[i].FilePath,
					queryResult,
					c.spec.Value)
			}
		}

		if !changed || dryRun {
			continue
		}

		// Update the target file with the new value
		switch len(c.spec.Query) > 0 {
		case true:
			err = c.contents[i].PutMultiple(c.spec.Query, c.spec.Value)

			if err != nil {
				return false, err
			}

		case false:
			err = c.contents[i].Put(c.spec.Key, c.spec.Value)

			if err != nil {
				return false, err
			}
		}

		err = c.contents[i].Write()
		if err != nil {
			return changed, err
		}
	}

	return changed, err
}

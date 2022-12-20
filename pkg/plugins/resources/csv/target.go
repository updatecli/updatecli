package csv

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (c *CSV) Target(source string, dryRun bool) (changed bool, err error) {

	changed, _, _, err = c.TargetFromSCM(source, nil, dryRun)
	if err != nil {
		return changed, err
	}

	return changed, err
}

// TargetFromSCM updates a scm repository based on the modified yaml file.
func (c *CSV) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (changed bool, files []string, message string, err error) {

	rootDir := ""
	if scm != nil {
		rootDir = scm.GetDirectory()
	}

	for i := range c.contents {
		filename := c.contents[i].FilePath

		// Target doesn't support updating files on remote http location
		if strings.HasPrefix(filename, "https://") ||
			strings.HasPrefix(filename, "http://") {
			return false, files, message, fmt.Errorf("URL scheme is not supported for CSV target: %q", c.spec.File)
		}

		if err := c.contents[i].Read(rootDir); err != nil {
			return false, files, message, fmt.Errorf("file %q does not exist", c.contents[i].FilePath)
		}

		if len(c.spec.Value) == 0 {
			c.spec.Value = source
		}

		resourceFile := c.contents[i].FilePath

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
				return false, files, message, err
			}

		case false:
			queryResult, err := c.contents[i].Query(c.spec.Key)

			if err != nil {
				return false, files, message, err
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
				return false, files, message, err
			}

		case false:
			err = c.contents[i].Put(c.spec.Key, c.spec.Value)

			if err != nil {
				return false, files, message, err
			}
		}

		err = c.contents[i].Write()
		if err != nil {
			return changed, files, message, err
		}

		files = append(files, resourceFile)
		message = fmt.Sprintf("Update key %q from file %q", c.spec.Key, c.spec.File)
	}

	return changed, files, message, err

}

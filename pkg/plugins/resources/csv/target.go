package csv

import (
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Target updates a scm repository based on the modified yaml file.
func (c *CSV) Target(source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {

	rootDir := ""
	if scm != nil {
		rootDir = scm.GetDirectory()
	}

	newValue := source
	if c.spec.Value != "" {
		newValue = c.spec.Value
	}

	resultTarget.Result = result.SUCCESS
	for i := range c.contents {
		filename := c.contents[i].FilePath

		// Target doesn't support updating files on remote http location
		if strings.HasPrefix(filename, "https://") ||
			strings.HasPrefix(filename, "http://") {
			return fmt.Errorf("URL scheme is not supported for CSV target: %q", c.spec.File)
		}

		if err := c.contents[i].Read(rootDir); err != nil {
			return fmt.Errorf("file %q does not exist", c.contents[i].FilePath)
		}

		resourceFile := c.contents[i].FilePath

		var queryResults []string
		var err error

		query := ""
		switch len(c.spec.Query) > 0 {
		case true:
			query = c.spec.Query
			queryResults, err = c.contents[i].MultipleQuery(c.spec.Query)

			if err != nil {
				return err
			}

		case false:
			query = c.spec.Key
			queryResult, err := c.contents[i].Query(c.spec.Key)

			if err != nil {
				return err
			}

			queryResults = append(queryResults, queryResult)

		}

		fileChanged := false
		for _, queryResult := range queryResults {
			resultTarget.Information = queryResult
			resultTarget.NewInformation = newValue

			switch resultTarget.NewInformation == resultTarget.Information {
			case true:
				resultTarget.Description = fmt.Sprintf("%s \n * Query %q correctly return %q from file %q",
					resultTarget.Description,
					query,
					resultTarget.NewInformation,
					c.contents[i].FilePath)

			case false:
				fileChanged = true
				resultTarget.Changed = true
				resultTarget.Files = append(resultTarget.Files, resourceFile)
				resultTarget.Result = result.ATTENTION
				resultTarget.Description = fmt.Sprintf("%s\n * Query %q, return update from %q to %q in file %q",
					resultTarget.Description,
					query,
					resultTarget.Information,
					resultTarget.NewInformation,
					c.contents[i].FilePath,
				)

			}
		}

		if !fileChanged || dryRun {
			continue
		}

		// Update the target file with the new value
		switch len(c.spec.Query) > 0 {
		case true:
			err = c.contents[i].PutMultiple(c.spec.Query, c.spec.Value)

			if err != nil {
				return err
			}

		case false:
			err = c.contents[i].Put(c.spec.Key, c.spec.Value)

			if err != nil {
				return err
			}
		}

		err = c.contents[i].Write()
		if err != nil {
			return err
		}
	}

	resultTarget.Description = strings.TrimPrefix(resultTarget.Description, "\n")

	if !dryRun {
		resultTarget.Description = strings.ReplaceAll(resultTarget.Description, "should be", "")
	}

	return nil

}

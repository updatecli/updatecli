package csv

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Target updates a scm repository based on the modified yaml file.
func (c *CSV) Target(_ context.Context, source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {

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
		switch c.engine {
		case ENGINEDASEL_V1:
			logrus.Debugf("Using engine %q", c.engine)
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

		case ENGINEDASEL_V2:
			logrus.Debugf("Using engine %q", c.engine)
			query = c.spec.Key
			queryResults, err = c.contents[i].QueryV2(c.spec.Key)
			if err != nil {
				return err
			}

		case ENGINEDASEL_V3:
			logrus.Debugf("Using engine %q", c.engine)
			query = c.spec.Key
			queryResults, err = c.contents[i].QueryV3(c.spec.Key)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("engine %q is not supported", c.engine)
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

		// Update the target file with the new value.
		// The dasel v3 engine mutates the native parsed rows in place; the CSV
		// writer then serializes those rows. v1 and v2 use the shared v1 node.
		switch c.engine {
		case ENGINEDASEL_V3:
			if err = c.contents[i].PutV3(c.spec.Key, c.spec.Value); err != nil {
				return err
			}

		default:
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

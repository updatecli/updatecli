package json

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Target updates a scm repository based on the modified yaml file.
func (j *Json) Target(source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {

	rootDir := ""
	if scm != nil {
		rootDir = scm.GetDirectory()
	}

	if len(j.spec.Value) == 0 {
		j.spec.Value = source
	}

	shouldMessage := ""
	if dryRun {
		shouldMessage = "should be "
	}

	for i := range j.contents {
		filename := j.contents[i].FilePath

		// Target doesn't support updating files on remote http location
		if strings.HasPrefix(filename, "https://") ||
			strings.HasPrefix(filename, "http://") {
			return fmt.Errorf("URL scheme is not supported for Json target: %q", j.spec.File)
		}

		if err := j.contents[i].Read(rootDir); err != nil {
			return fmt.Errorf("file %q does not exist", j.contents[i].FilePath)
		}

		resourceFile := j.contents[i].FilePath

		var queryResults []string
		var err error

		switch len(j.spec.Query) > 0 {
		case true:
			queryResults, err = j.contents[i].MultipleQuery(j.spec.Query)

			if err != nil {
				return err
			}

		case false:
			queryResult, err := j.contents[i].Query(j.spec.Key)

			if err != nil {
				return err
			}

			queryResults = append(queryResults, queryResult)

		}

		for _, queryResult := range queryResults {
			resultTarget.Information = queryResult

			switch queryResult == j.spec.Value {
			case true:
				resultTarget.Information = queryResult
				resultTarget.NewInformation = queryResult
				resultTarget.Result = result.SUCCESS
				resultTarget.Changed = false

				logrus.Infof("%s\nkey %q, from file %q, is correctly set to %q",
					resultTarget.Description,
					j.spec.Key,
					j.contents[i].FilePath,
					resultTarget.NewInformation)

			case false:
				resultTarget.Result = result.ATTENTION
				resultTarget.Changed = true
				resultTarget.NewInformation = j.spec.Value
				resultTarget.Information = queryResult

				logrus.Infof("%s\nkey %q, from file %q, %supdated from %q to %q",
					resultTarget.Description,
					j.spec.Key,
					shouldMessage,
					j.contents[i].FilePath,
					resultTarget.Information,
					resultTarget.NewInformation)
			}
		}

		if !resultTarget.Changed || dryRun {
			continue
		}

		// Update the target file with the new value
		switch len(j.spec.Query) > 0 {
		case true:
			err = j.contents[i].PutMultiple(j.spec.Query, j.spec.Value)

			if err != nil {
				return err
			}

		case false:
			err = j.contents[i].Put(j.spec.Key, j.spec.Value)

			if err != nil {
				return err
			}
		}

		err = j.contents[i].Write()
		if err != nil {
			return err
		}

		resultTarget.Files = append(resultTarget.Files, resourceFile)
	}

	return nil

}

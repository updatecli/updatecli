package json

import (
	"fmt"
	"slices"
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

	unModifiedDescriptions := []string{}

	modifiedDescriptions := []string{}
	modifiedFiles := []string{}
	modifiedValues := []string{}

	for i := range j.contents {
		filename := j.contents[i].FilePath

		// Target doesn't support updating files on remote http location
		if strings.HasPrefix(filename, "https://") ||
			strings.HasPrefix(filename, "http://") {
			return fmt.Errorf("URL scheme is not supported for Json target: %q", j.spec.File)
		}

		if err := j.contents[i].Read(rootDir); err != nil {
			return fmt.Errorf("loading json file %q: %w", filename, err)
		}

		var queryResults []string
		var err error

		switch j.engine {
		case ENGINEDASEL_V1:
			logrus.Debugf("Using engine %q", j.engine)
			switch len(j.spec.Query) > 0 {
			case true:
				queryResults, err = j.contents[i].MultipleQuery(j.spec.Query)

				if err != nil {
					return fmt.Errorf("querying json file %q: %w", filename, err)
				}

			case false:
				queryResult, err := j.contents[i].Query(j.spec.Key)

				if err != nil {
					return fmt.Errorf("querying json file %q: %w", filename, err)
				}

				queryResults = append(queryResults, queryResult)

			}

		case ENGINEDASEL_V2:
			logrus.Debugf("Using engine %q", ENGINEDASEL_V2)
			queryResults, err = j.contents[i].QueryV2(j.spec.Key)
			if err != nil {
				return fmt.Errorf("querying file %q: %w", j.contents[i].FilePath, err)
			}

		default:
			return fmt.Errorf("engine %q is not supported", j.engine)
		}

		resultChanged := false
		for _, queryResult := range queryResults {
			resultTarget.Information = queryResult

			switch queryResult == j.spec.Value {
			case true:
				description := fmt.Sprintf("key %q, from file %q, is correctly set to %q",
					j.spec.Key,
					filename,
					j.spec.Value)

				if !slices.Contains(unModifiedDescriptions, description) {
					unModifiedDescriptions = append(unModifiedDescriptions, description)
				}

			case false:
				resultChanged = true
				if !slices.Contains(modifiedFiles, filename) {
					modifiedFiles = append(modifiedFiles, filename)
				}

				if !slices.Contains(modifiedValues, resultTarget.Information) {
					modifiedValues = append(modifiedValues, resultTarget.Information)
				}

				description := fmt.Sprintf("key %q, from file %q, %s updated from %q to %q",
					j.spec.Key,
					filename,
					shouldMessage,
					queryResult,
					j.spec.Value)

				if !slices.Contains(modifiedDescriptions, description) {
					modifiedDescriptions = append(modifiedDescriptions, description)
				}
			}
		}

		if !resultChanged || dryRun {
			continue
		}

		// Update the target file with the new value
		switch len(j.spec.Query) > 0 {
		case true:
			err = j.contents[i].PutMultiple(j.spec.Query, j.spec.Value)

			if err != nil {
				return fmt.Errorf("updating json file %q: %w", filename, err)
			}

		case false:
			err = j.contents[i].Put(j.spec.Key, j.spec.Value)

			if err != nil {
				return fmt.Errorf("updating json file %q: %w", filename, err)
			}
		}

		err = j.contents[i].Write()
		if err != nil {
			return fmt.Errorf("writing json file %q: %w", filename, err)
		}
	}

	if len(modifiedDescriptions) == 0 && len(unModifiedDescriptions) > 0 {
		resultTarget.Result = result.SUCCESS
		resultTarget.Changed = false
		resultTarget.NewInformation = j.spec.Value
		resultTarget.Information = j.spec.Value
		resultTarget.Description = fmt.Sprintf(
			"all json file(s) up to date:\n\t* %s\n",
			strings.Join(unModifiedDescriptions, "\n\t*"))

	} else if len(modifiedDescriptions) > 0 {
		resultTarget.Files = modifiedFiles
		resultTarget.Result = result.ATTENTION
		resultTarget.Changed = true
		resultTarget.NewInformation = j.spec.Value
		resultTarget.Information = fmt.Sprintf("%v", modifiedValues)
		resultTarget.Description = fmt.Sprintf(
			"%d json file(s) updated:\n\t* %s\n",
			len(modifiedDescriptions), strings.Join(modifiedDescriptions, "\n\t*"))

	} else if len(modifiedDescriptions) == 0 && len(unModifiedDescriptions) == 0 {
		resultTarget.Result = result.SKIPPED
		resultTarget.Description = "no json file(s) found"

	} else {
		description := "unable to determine the result"
		resultTarget.Result = result.FAILURE
		resultTarget.Description = description
		return fmt.Errorf("%s", description)
	}

	return nil

}

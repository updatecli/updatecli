package json

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (j *Json) Target(source string, dryRun bool) (changed bool, err error) {

	changed, _, _, err = j.TargetFromSCM(source, nil, dryRun)
	if err != nil {
		return changed, err
	}

	return changed, err
}

// TargetFromSCM updates a scm repository based on the modified yaml file.
func (j *Json) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (changed bool, files []string, message string, err error) {

	rootDir := ""
	if scm != nil {
		rootDir = scm.GetDirectory()
	}

	for i := range j.contents {
		filename := j.contents[i].FilePath

		// Target doesn't support updating files on remote http location
		if strings.HasPrefix(filename, "https://") ||
			strings.HasPrefix(filename, "http://") {
			return false, files, message, fmt.Errorf("URL scheme is not supported for Json target: %q", j.spec.File)
		}

		if err := j.contents[i].Read(rootDir); err != nil {
			return false, files, message, fmt.Errorf("file %q does not exist", j.contents[i].FilePath)
		}

		if len(j.spec.Value) == 0 {
			j.spec.Value = source
		}

		resourceFile := j.contents[i].FilePath

		// Override value from source if not yet defined
		if len(j.spec.Value) == 0 {
			j.spec.Value = source
		}

		var queryResults []string
		var err error

		switch len(j.spec.Query) > 0 {
		case true:
			queryResults, err = j.contents[i].MultipleQuery(j.spec.Query)

			if err != nil {
				return false, files, message, err
			}

		case false:
			queryResult, err := j.contents[i].Query(j.spec.Key)

			if err != nil {
				return false, files, message, err
			}

			queryResults = append(queryResults, queryResult)

		}

		for _, queryResult := range queryResults {
			switch queryResult == j.spec.Value {
			case true:
				logrus.Infof("%s Key '%s', from file '%v', is correctly set to %s'",
					result.SUCCESS,
					j.spec.Key,
					j.contents[i].FilePath,
					j.spec.Value)

			case false:
				changed = true
				logrus.Infof("%s Key '%s', from file '%v', is incorrectly set to %s and should be %s'",
					result.ATTENTION,
					j.spec.Key,
					j.contents[i].FilePath,
					queryResult,
					j.spec.Value)
			}
		}

		if !changed || dryRun {
			continue
		}

		// Update the target file with the new value
		switch len(j.spec.Query) > 0 {
		case true:
			err = j.contents[i].PutMultiple(j.spec.Query, j.spec.Value)

			if err != nil {
				return false, files, message, err
			}

		case false:
			err = j.contents[i].Put(j.spec.Key, j.spec.Value)

			if err != nil {
				return false, files, message, err
			}
		}

		err = j.contents[i].Write()
		if err != nil {
			return changed, files, message, err
		}

		files = append(files, resourceFile)
		message = fmt.Sprintf("Update key %q from file %q", j.spec.Key, j.spec.File)
	}

	return changed, files, message, err

}

package toml

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (t *Toml) Target(source string, dryRun bool) (changed bool, err error) {

	changed, _, _, err = t.TargetFromSCM(source, nil, dryRun)
	if err != nil {
		return changed, err
	}

	return changed, err
}

// TargetFromSCM updates a scm repository based on the modified yaml file.
func (t *Toml) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (changed bool, files []string, message string, err error) {

	rootDir := ""
	if scm != nil {
		rootDir = scm.GetDirectory()
	}

	for i := range t.contents {
		filename := t.contents[i].FilePath

		// Target doesn't support updating files on remote http location
		if strings.HasPrefix(filename, "https://") ||
			strings.HasPrefix(filename, "http://") {
			return false, files, message, fmt.Errorf("URL scheme is not supported for Json target: %q", t.spec.File)
		}

		if err := t.contents[i].Read(rootDir); err != nil {
			return false, files, message, fmt.Errorf("file %q does not exist", t.contents[i].FilePath)
		}

		if len(t.spec.Value) == 0 {
			t.spec.Value = source
		}

		resourceFile := t.contents[i].FilePath

		// Override value from source if not yet defined
		if len(t.spec.Value) == 0 {
			t.spec.Value = source
		}

		var queryResults []string
		var err error

		switch t.spec.Multiple {
		case true:
			queryResults, err = t.contents[i].MultipleQuery(t.spec.Key)

			if err != nil {
				return false, files, message, err
			}

		case false:
			queryResult, err := t.contents[i].Query(t.spec.Key)

			if err != nil {
				return false, files, message, err
			}

			queryResults = append(queryResults, queryResult)

		}

		for _, queryResult := range queryResults {
			switch queryResult == t.spec.Value {
			case true:
				logrus.Infof("%s Key '%s', from file '%v', is correctly set to %s'",
					result.SUCCESS,
					t.spec.Key,
					t.contents[i].FilePath,
					t.spec.Value)

			case false:
				changed = true
				logrus.Infof("%s Key '%s', from file '%v', is incorrectly set to %s and should be %s'",
					result.ATTENTION,
					t.spec.Key,
					t.contents[i].FilePath,
					queryResult,
					t.spec.Value)
			}
		}

		if !changed || dryRun {
			continue
		}

		// Update the target file with the new value
		switch t.spec.Multiple {
		case true:
			err = t.contents[i].PutMultiple(t.spec.Key, t.spec.Value)

			if err != nil {
				return false, files, message, err
			}

		case false:
			err = t.contents[i].Put(t.spec.Key, t.spec.Value)

			if err != nil {
				return false, files, message, err
			}
		}

		err = t.contents[i].Write()
		if err != nil {
			return changed, files, message, err
		}

		files = append(files, resourceFile)
		message = fmt.Sprintf("Update key %q from file %q", t.spec.Key, t.spec.File)
	}

	return changed, files, message, err

}

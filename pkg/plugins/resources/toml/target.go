package toml

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/result"
)

func (t *Toml) Target(source string, dryRun bool) (changed bool, err error) {
	rootDir := ""

	for i := range t.contents {
		filename := t.contents[i].FilePath

		// Target doesn't support updating files on remote http location
		if strings.HasPrefix(filename, "https://") ||
			strings.HasPrefix(filename, "http://") {
			return false, fmt.Errorf("URL scheme is not supported for Json target: %q", t.spec.File)
		}

		if err := t.contents[i].Read(rootDir); err != nil {
			return false, fmt.Errorf("file %q does not exist", t.contents[i].FilePath)
		}

		if len(t.spec.Value) == 0 {
			t.spec.Value = source
		}

		// Override value from source if not yet defined
		if len(t.spec.Value) == 0 {
			t.spec.Value = source
		}

		var queryResults []string
		var err error

		switch len(t.spec.Query) > 0 {
		case true:
			queryResults, err = t.contents[i].MultipleQuery(t.spec.Query)

			if err != nil {
				return false, err
			}

		case false:
			queryResult, err := t.contents[i].Query(t.spec.Key)

			if err != nil {
				return false, err
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
		switch len(t.spec.Query) > 0 {
		case true:
			err = t.contents[i].PutMultiple(t.spec.Query, t.spec.Value)

			if err != nil {
				return false, err
			}

		case false:
			err = t.contents[i].Put(t.spec.Key, t.spec.Value)

			if err != nil {
				return false, err
			}
		}

		err = t.contents[i].Write()
		if err != nil {
			return changed, err
		}
	}

	return changed, err
}

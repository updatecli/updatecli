package toml

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Target updates a scm repository based on the modified yaml file.
func (t *Toml) Target(_ context.Context, source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {

	rootDir := ""
	if scm != nil {
		rootDir = scm.GetDirectory()
	}

	for i := range t.contents {
		filename := t.contents[i].FilePath

		// Target doesn't support updating files on remote http location
		if strings.HasPrefix(filename, "https://") ||
			strings.HasPrefix(filename, "http://") {
			return fmt.Errorf("URL scheme is not supported for Toml target: %q", t.spec.File)
		}

		if err := t.contents[i].Read(rootDir); err != nil {
			return fmt.Errorf("file %q does not exist", t.contents[i].FilePath)
		}

		if len(t.spec.Value) == 0 {
			t.spec.Value = source
		}
		resultTarget.NewInformation = t.spec.Value

		resourceFile := t.contents[i].FilePath

		// Override value from source if not yet defined
		if len(t.spec.Value) == 0 {
			t.spec.Value = source
		}

		var queryResults []string
		var err error

		switch t.engine {
		case ENGINEDASEL_V1:
			logrus.Debugf("Using engine %q", t.engine)
			switch len(t.spec.Query) > 0 {
			case true:
				queryResults, err = t.contents[i].MultipleQuery(t.spec.Query)
				if err != nil {
					return err
				}

			case false:
				if t.spec.CreateMissingKey {
					queryResults, err = t.contents[i].MultipleQuery(t.spec.Query)
					if err != nil {
						return err
					}
				} else {
					queryResult, err := t.contents[i].Query(t.spec.Key)
					if err != nil {
						return err
					}
					queryResults = append(queryResults, queryResult)
				}

			}

		case ENGINEDASEL_V2:
			logrus.Debugf("Using engine %q", t.engine)
			queryResults, err = t.contents[i].QueryV2(t.spec.Key)
			if err != nil {
				return fmt.Errorf("querying file %q: %w", t.contents[i].FilePath, err)
			}

		case ENGINEDASEL_V3:
			logrus.Debugf("Using engine %q", t.engine)
			queryResults, err = t.contents[i].QueryV3(t.spec.Key)
			if err != nil {
				return fmt.Errorf("querying file %q: %w", t.contents[i].FilePath, err)
			}

		default:
			return fmt.Errorf("engine %q is not supported", t.engine)
		}

		changedFile := false
		for _, queryResult := range queryResults {
			switch queryResult == resultTarget.NewInformation {
			case true:
				resultTarget.Description = fmt.Sprintf("%s\nkey %q, from file %q, is correctly set to %q",
					resultTarget.Description,
					t.spec.Key,
					t.contents[i].FilePath,
					t.spec.Value)

			case false:
				changedFile = true
				resultTarget.Information = queryResult
				resultTarget.Result = result.ATTENTION
				resultTarget.Changed = true
				resultTarget.Description = fmt.Sprintf("%s\nkey %q, from file %q, is incorrectly set to %q and should be %q",
					resultTarget.Description,
					t.spec.Key,
					t.contents[i].FilePath,
					queryResult,
					t.spec.Value)
			}
		}

		if !changedFile || dryRun {
			continue
		}

		// Update the target file with the new value.
		// The dasel v3 engine operates on the native parsed data and requires its
		// own put/write path; v1 and v2 both write through the shared v1 node.
		switch t.engine {
		case ENGINEDASEL_V3:
			if err = t.contents[i].PutV3(t.spec.Key, t.spec.Value); err != nil {
				return err
			}

			if err = t.contents[i].WriteV3(); err != nil {
				return err
			}

		default:
			switch len(t.spec.Query) > 0 {
			case true:
				err = t.contents[i].PutMultiple(t.spec.Query, t.spec.Value)

				if err != nil {
					return err
				}

			case false:
				err = t.contents[i].Put(t.spec.Key, t.spec.Value)

				if err != nil {
					return err
				}
			}

			err = t.contents[i].Write()
			if err != nil {
				return err
			}
		}

		resultTarget.Files = append(resultTarget.Files, resourceFile)
	}

	resultTarget.Description = strings.TrimPrefix(resultTarget.Description, "\n")

	return nil

}

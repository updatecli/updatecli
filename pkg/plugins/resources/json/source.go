package json

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

var (
	ErrSpecVersionFilterRequireMultiple = errors.New("in the context of a source, parameter \"versionfilter\" and \"query\" must be used together")
)

func (j *Json) Source(workingDir string, resultSource *result.Source) error {

	if len(j.contents) > 1 {
		return errors.New("source only supports one file")
	}

	if j.engine != ENGINEDASEL_V2 &&
		((len(j.spec.Query) > 0 && j.spec.VersionFilter.IsZero()) ||
			(len(j.spec.Query) == 0 && !j.spec.VersionFilter.IsZero())) {
		return ErrSpecVersionFilterRequireMultiple
	}

	content := j.contents[0]

	sourceOutput := ""
	if err := content.Read(workingDir); err != nil {
		return fmt.Errorf("reading json file: %w", err)
	}

	query := ""
	switch j.engine {
	case ENGINEDASEL_V1:
		logrus.Debugf("Using engine %q", j.engine)
		switch len(j.spec.Query) > 0 {
		case true:
			query = j.spec.Query
			queryResults, err := content.MultipleQuery(query)

			if err != nil {
				return fmt.Errorf("running multiple query: %w", err)
			}

			j.foundVersion, err = j.versionFilter.Search(queryResults)
			if err != nil {
				return fmt.Errorf("filtering information: %w", err)
			}
			sourceOutput = j.foundVersion.GetVersion()

		case false:
			query = j.spec.Key
			queryResult, err := content.DaselNode.Query(query)
			if err != nil {
				// Catch error message returned by Dasel, if it couldn't find the node
				// This is approach is not very robust
				// https://github.com/TomWright/dasel/blob/master/node_query.go#L58

				if strings.HasPrefix(err.Error(), "could not find value:") {
					err := fmt.Errorf("%s cannot find value for path %q from file %q",
						result.FAILURE,
						j.spec.Key,
						content.FilePath)
					return err
				}
				return err
			}

			sourceOutput = queryResult.String()
		}
	case ENGINEDASEL_V2:
		logrus.Debugf("Using engine %q", j.engine)
		queryResults, err := content.QueryV2(j.spec.Key)

		if err != nil {
			if strings.Contains(err.Error(), "property not found") {
				err := fmt.Errorf("%s cannot find value for path %q from file %q",
					result.FAILURE,
					j.spec.Key,
					content.FilePath)
				return err
			}
			return fmt.Errorf("running query %q: %w", j.spec.Key, err)
		}

		j.foundVersion, err = j.versionFilter.Search(queryResults)
		if err != nil {
			return fmt.Errorf("filtering information: %w", err)
		}
		sourceOutput = j.foundVersion.GetVersion()

	default:
		return fmt.Errorf("engine %q not supported", j.engine)
	}

	resultSource.Information = sourceOutput
	resultSource.Result = result.SUCCESS
	resultSource.Description = fmt.Sprintf("value %q, found in file %q, for key %q",
		sourceOutput,
		content.FilePath,
		query)

	return nil
}

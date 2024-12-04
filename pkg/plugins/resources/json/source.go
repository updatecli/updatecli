package json

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/result"
)

var (
	ErrSpecVersionFilterRequireMultiple = errors.New("in the context of a source, parameter \"versionfilter\" and \"query\" must be used together")
)

func (j *Json) Source(workingDir string, resultSource *result.Source) error {

	if len(j.contents) > 1 {
		return errors.New("source only supports one file")
	}

	if (len(j.spec.Query) > 0 && j.spec.VersionFilter.IsZero() && !j.isList) ||
		(len(j.spec.Query) == 0) && !j.spec.VersionFilter.IsZero() {
		return ErrSpecVersionFilterRequireMultiple
	}

	content := j.contents[0]

	sourceOutputs := []result.SourceInformation{}

	if err := content.Read(workingDir); err != nil {
		return fmt.Errorf("reading json file: %w", err)
	}

	query := ""
	switch len(j.spec.Query) > 0 {
	case true:
		query = j.spec.Query
		queryResults, err := content.MultipleQuery(query)

		if err != nil {
			return fmt.Errorf("running multiple query: %w", err)
		}

		if j.isList {
			for i, queryResult := range queryResults {
				sourceOutputs = append(sourceOutputs, result.SourceInformation{Key: strconv.Itoa(i), Value: queryResult})
			}
		} else {

			j.foundVersion, err = j.versionFilter.Search(queryResults)
			if err != nil {
				return fmt.Errorf("filtering information: %w", err)
			}
			sourceOutputs = []result.SourceInformation{{Value: j.foundVersion.GetVersion()}}
		}

	case false:
		if j.isList {
			err := fmt.Errorf("%s json source can only be configured as a list with the `query` param.", result.FAILURE)
			return err
		}
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

		sourceOutputs = []result.SourceInformation{{Value: queryResult.String()}}
	}

	resultSource.Information = sourceOutputs
	resultSource.Result = result.SUCCESS
	sourceOutputsValues := make([]string, len(sourceOutputs))
	for i, output := range sourceOutputs {
		sourceOutputsValues[i] = output.Value
	}

	resultSource.Description = fmt.Sprintf("values %q, found in file %q, for key %q",
		strings.Join(sourceOutputsValues, " "),
		content.FilePath,
		query)

	return nil
}

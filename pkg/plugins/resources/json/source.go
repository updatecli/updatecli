package json

import (
	"errors"
	"fmt"
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

	if (len(j.spec.Query) > 0 && j.spec.VersionFilter.IsZero()) ||
		(len(j.spec.Query) == 0) && !j.spec.VersionFilter.IsZero() {
		return ErrSpecVersionFilterRequireMultiple
	}

	content := j.contents[0]

	sourceOutput := ""
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

	resultSource.Information = []result.SourceInformation{{
		Key:   resultSource.ID,
		Value: sourceOutput,
	}}
	resultSource.Result = result.SUCCESS
	resultSource.Description = fmt.Sprintf("value %q, found in file %q, for key %q",
		sourceOutput,
		content.FilePath,
		query)

	return nil
}

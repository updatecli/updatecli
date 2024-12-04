package toml

import (
	"errors"
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/result"
)

var (
	ErrSpecVersionFilterRequireMultiple = errors.New("in the context of a source, parameter \"versionfilter\" and \"query\" must be used together")
)

func (t *Toml) Source(workingDir string, resultSource *result.Source) error {

	if len(t.contents) > 1 {
		return errors.New("source only supports one file")
	}

	if (len(t.spec.Query) > 0 && t.spec.VersionFilter.IsZero()) ||
		(len(t.spec.Query) == 0) && !t.spec.VersionFilter.IsZero() {
		return ErrSpecVersionFilterRequireMultiple
	}

	content := t.contents[0]

	sourceOutput := ""

	if err := content.Read(workingDir); err != nil {
		return fmt.Errorf("reading toml file: %w", err)
	}

	query := ""
	switch len(t.spec.Query) > 0 {
	case true:
		query = t.spec.Query
		queryResults, err := content.MultipleQuery(query)

		if err != nil {
			return fmt.Errorf("running multiple query: %w", err)
		}

		t.foundVersion, err = t.versionFilter.Search(queryResults)
		if err != nil {
			return fmt.Errorf("filtering result: %w", err)
		}
		sourceOutput = t.foundVersion.GetVersion()

	case false:
		query = t.spec.Key
		queryResult, err := content.DaselNode.Query(query)
		if err != nil {
			// Catch error message returned by Dasel, if it couldn't find the node
			// This is approach is not very robust
			// https://github.com/TomWright/dasel/blob/master/node_query.go#L58

			if strings.HasPrefix(err.Error(), "could not find value:") {
				return fmt.Errorf("cannot find value for path %q from file %q",
					t.spec.Key,
					content.FilePath)
			}
			return fmt.Errorf("running query: %w", err)
		}

		sourceOutput = queryResult.String()
	}

	resultSource.Information = []result.SourceInformation{{
		Key:   resultSource.ID,
		Value: sourceOutput,
	}}
	resultSource.Result = result.SUCCESS
	resultSource.Description = fmt.Sprintf("value %q, found in file %q, for key %q'",
		sourceOutput,
		content.FilePath,
		query)

	return nil
}

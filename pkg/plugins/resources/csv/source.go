package csv

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

var (
	ErrSpecVersionFilterRequireMultiple = errors.New("in the context of a source, parameter \"versionfilter\" and \"query\" must be used together")
)

func (c *CSV) Source(_ context.Context, workingDir string, resultSource *result.Source) error {

	if len(c.contents) > 1 {
		return errors.New("source only supports one file")
	}

	if c.engine != ENGINEDASEL_V2 && c.engine != ENGINEDASEL_V3 &&
		((len(c.spec.Query) > 0 && c.spec.VersionFilter.IsZero()) ||
			(len(c.spec.Query) == 0) && !c.spec.VersionFilter.IsZero()) {
		return ErrSpecVersionFilterRequireMultiple
	}

	content := c.contents[0]

	sourceOutput := ""

	if err := content.Read(workingDir); err != nil {
		return fmt.Errorf("reading csv file: %w", err)
	}

	query := ""
	switch c.engine {
	case ENGINEDASEL_V1:
		logrus.Debugf("Using engine %q", c.engine)
		switch len(c.spec.Query) > 0 {
		case true:
			query = c.spec.Query
			queryResults, err := content.MultipleQuery(query)

			if err != nil {
				return fmt.Errorf("running multi query: %w", err)
			}

			c.foundVersion, err = c.versionFilter.Search(queryResults)
			if err != nil {
				return fmt.Errorf("filtering version: %w", err)
			}
			sourceOutput = c.foundVersion.GetVersion()

		case false:
			query = c.spec.Key
			queryResult, err := content.DaselNode.Query(query)
			if err != nil {
				// Catch error message returned by Dasel, if it couldn't find the node
				// This is approach is not very robust
				// https://github.com/TomWright/dasel/blob/master/node_query.go#L58

				if strings.HasPrefix(err.Error(), "could not find value:") {
					err := fmt.Errorf("cannot find value for path %q from file %q",
						c.spec.Key,
						content.FilePath)
					return err
				}
				return err
			}

			sourceOutput = queryResult.String()
		}

	case ENGINEDASEL_V2:
		logrus.Debugf("Using engine %q", c.engine)
		query = c.spec.Key
		queryResults, err := content.QueryV2(c.spec.Key)
		if err != nil {
			if strings.Contains(err.Error(), "property not found") {
				return fmt.Errorf("cannot find value for path %q from file %q",
					c.spec.Key,
					content.FilePath)
			}
			return fmt.Errorf("running query %q: %w", c.spec.Key, err)
		}

		c.foundVersion, err = c.versionFilter.Search(queryResults)
		if err != nil {
			return fmt.Errorf("filtering version: %w", err)
		}
		sourceOutput = c.foundVersion.GetVersion()

	case ENGINEDASEL_V3:
		logrus.Debugf("Using engine %q", c.engine)
		query = c.spec.Key
		queryResults, err := content.QueryV3(c.spec.Key)
		if err != nil {
			if strings.Contains(err.Error(), "map key not found") {
				return fmt.Errorf("cannot find value for path %q from file %q",
					c.spec.Key,
					content.FilePath)
			}
			return fmt.Errorf("running query %q: %w", c.spec.Key, err)
		}

		c.foundVersion, err = c.versionFilter.Search(queryResults)
		if err != nil {
			return fmt.Errorf("filtering version: %w", err)
		}
		sourceOutput = c.foundVersion.GetVersion()

	default:
		return fmt.Errorf("engine %q not supported", c.engine)
	}

	resultSource.Result = result.SUCCESS
	resultSource.Information = sourceOutput
	resultSource.Description = fmt.Sprintf("csv value %q, found in file %q, for path %q",
		sourceOutput,
		content.FilePath,
		query)

	return nil

}

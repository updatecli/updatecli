package dasel

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// MultipleQuery returns multiple query from a Dasel Node
func (f *FileContent) MultipleQuery(query string) ([]string, error) {
	var results []string

	if f.DaselNode == nil {
		return []string{}, ErrDaselFailedParsingByteFormat
	}

	queryResults, err := f.DaselNode.QueryMultiple(query)
	if err != nil {
		// Catch error message returned by Dasel, if it couldn't find the node
		// This is approach is not very robust
		// https://github.com/TomWright/dasel/blob/master/node_query.go#L58

		if strings.HasPrefix(err.Error(), "could not find multiple value:") {
			logrus.Debugln(err)
			err = fmt.Errorf("could not find multiple value for query %q from file %q",
				query,
				f.FilePath)

			return []string{}, err
		}

		return []string{}, err
	}

	if queryResults == nil {
		err = fmt.Errorf("could not find multiple value for query %q from file %q",
			query,
			f.FilePath)
		return []string{}, err
	}

	for i := range queryResults {

		results = append(results, queryResults[i].String())
	}

	return results, nil

}

// Query returns the value for a specific query from a Dasel node
func (f *FileContent) Query(query string) (string, error) {
	if f.DaselNode == nil {
		return "", ErrDaselFailedParsingByteFormat
	}

	queryResult, err := f.DaselNode.Query(query)
	if err != nil {
		// Catch error message returned by Dasel, if it couldn't find the node
		// This is approach is not very robust
		// https://github.com/TomWright/dasel/blob/master/node_query.go#L58

		if strings.HasPrefix(err.Error(), "could not find value:") {
			logrus.Debugln(err)
			err = fmt.Errorf("could not find value for query %q from file %q",
				query,
				f.FilePath)
			return "", err
		}

		return "", err
	}

	if queryResult == nil {
		err = fmt.Errorf("could not find value for query %q from file %q",
			query,
			f.FilePath)
		return "", err
	}

	return queryResult.String(), nil

}

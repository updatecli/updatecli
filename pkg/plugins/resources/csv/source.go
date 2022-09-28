package csv

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/tomwright/dasel"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (c *CSV) Source(workingDir string) (string, error) {

	// Test at runtime if a file exist
	if !c.contentRetriever.FileExists(c.spec.File) {
		return "", fmt.Errorf("the CSV file %q does not exist", c.spec.File)
	}

	if err := c.Read(); err != nil {
		return "", err
	}

	if err := c.ReadFromFile(); err != nil {
		return "", err
	}

	rootNode := dasel.New(c.csvDocument.Documents())

	//rootNode, err := dasel.NewFromReader(strings.NewReader(c.currentContent), "csv")

	if rootNode == nil {
		return "", ErrDaselFailedParsingJSONByteFormat
	}

	queryResult, err := rootNode.Query(c.spec.Key)
	if err != nil {
		// Catch error message returned by Dasel, if it couldn't find the node
		// This is approach is not very robust
		// https://github.com/TomWright/dasel/blob/master/node_query.go#L58

		if strings.HasPrefix(err.Error(), "could not find value:") {
			logrus.Infof("%s cannot find value for path %q from file %q",
				result.FAILURE,
				c.spec.Key,
				c.spec.File)
			return "", nil
		}
		return "", err
	}

	logrus.Infof("%s Value %q, found in file %q, for key %q'",
		result.SUCCESS,
		queryResult.String(),
		c.spec.File,
		c.spec.Key)

	return queryResult.String(), nil
}

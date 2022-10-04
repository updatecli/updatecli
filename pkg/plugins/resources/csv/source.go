package csv

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (c *CSV) Source(workingDir string) (string, error) {

	if len(c.contents) > 1 {
		return "", errors.New("source only supports one file")
	}

	sourceOutput := ""
	for i := range c.contents {
		if err := c.contents[i].Read(workingDir); err != nil {
			return "", err
		}

		queryResult, err := c.contents[i].DaselNode.Query(c.spec.Key)
		if err != nil {
			// Catch error message returned by Dasel, if it couldn't find the node
			// This is approach is not very robust
			// https://github.com/TomWright/dasel/blob/master/node_query.go#L58

			if strings.HasPrefix(err.Error(), "could not find value:") {
				err := fmt.Errorf("could not find value for path %q from file %q",
					c.spec.Key,
					c.contents[i].FilePath)
				return "", err
			}
			return "", err
		}

		logrus.Infof("%s Value %q, found in file %q, for key %q'",
			result.SUCCESS,
			queryResult.String(),
			c.contents[i].FilePath,
			c.spec.Key)

		sourceOutput = queryResult.String()

	}

	return sourceOutput, nil

}

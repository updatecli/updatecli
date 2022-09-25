package csv

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/tomwright/dasel"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (c *CSV) Condition(source string) (bool, error) {
	return c.ConditionFromSCM(source, nil)
}

func (c *CSV) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	if scm != nil {
		c.spec.File = joinPathWithWorkingDirectoryPath(c.spec.File, scm.GetDirectory())
	}

	// Test at runtime if a file exist
	if !c.contentRetriever.FileExists(c.spec.File) {
		return false, fmt.Errorf("the CSV file %q does not exist", c.spec.File)
	}

	if err := c.Read(); err != nil {
		return false, err
	}

	// Override value from source if not yet defined
	if len(c.spec.Value) == 0 {
		c.spec.Value = source
	}

	if err := c.ReadFromFile(); err != nil {
		return false, err
	}

	rootNode := dasel.New(c.csvDocument.Documents())

	if rootNode == nil {
		return false, ErrDaselFailedParsingJSONByteFormat
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
			return false, nil
		}

		return false, err
	}

	if queryResult.String() == c.spec.Value {
		logrus.Infof("%s Key '%s', from file '%v', is correctly set to %s'",
			result.SUCCESS,
			c.spec.Key,
			c.spec.File,
			c.spec.Value)
		return true, nil
	} else {
		logrus.Infof("%s Key '%s', from file '%v', is incorrectly set to %s and should be %s'",
			result.ATTENTION,
			c.spec.Key,
			c.spec.File,
			queryResult.String(),
			c.spec.Value)
	}
	return false, nil
}

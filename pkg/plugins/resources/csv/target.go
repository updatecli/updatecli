package csv

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/tomwright/dasel"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (c *CSV) Target(source string, dryRun bool) (changed bool, err error) {

	changed, _, _, err = c.TargetFromSCM(source, nil, dryRun)
	if err != nil {
		return changed, err
	}

	return changed, err
}

// TargetFromSCM updates a scm repository based on the modified yaml file.
func (c *CSV) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (changed bool, files []string, message string, err error) {

	if strings.HasPrefix(c.spec.File, "https://") ||
		strings.HasPrefix(c.spec.File, "http://") {
		return false, files, message, fmt.Errorf("URL scheme is not supported for Json target: %q", c.spec.File)
	}

	if scm != nil {
		c.spec.File = joinPathWithWorkingDirectoryPath(c.spec.File, scm.GetDirectory())
	}

	// Test at runtime if a file exist
	if !c.contentRetriever.FileExists(c.spec.File) {
		return false, files, message, fmt.Errorf("the Json file %q does not exist", c.spec.File)
	}

	if len(c.spec.Value) == 0 {
		c.spec.Value = source
	}

	resourceFile := ""
	if scm != nil {
		resourceFile = filepath.Join(scm.GetDirectory(), c.spec.File)
	} else {
		resourceFile = c.spec.File
	}

	if err := c.Read(); err != nil {
		return false, []string{}, "", err
	}

	// Override value from source if not yet defined
	if len(c.spec.Value) == 0 {
		c.spec.Value = source
	}

	if err := c.ReadFromFile(); err != nil {
		return false, []string{}, "", err
	}

	rootNode := dasel.New(c.csvDocument.Documents())

	if rootNode == nil {
		return changed, files, message, ErrDaselFailedParsingJSONByteFormat
	}

	switch c.spec.Multiple {
	case true:
		changed, err = c.multipleTargetQuery(rootNode)
	case false:
		changed, err = c.singleTargetQuery(rootNode)
	}

	if err != nil {
		return false, files, message, err
	}

	if !changed {
		return changed, files, message, nil
	}

	if !dryRun {

		fileInfo, err := os.Stat(resourceFile)
		if err != nil {
			return changed, files, message, fmt.Errorf("[%s] unable to get file info: %w", c.spec.File, err)
		}

		logrus.Debugf("fileInfo for %s mode=%s", resourceFile, fileInfo.Mode().String())

		user, err := user.Current()
		if err != nil {
			logrus.Errorf("unable to get user info: %s", err)
		}

		logrus.Debugf("user: username=%s, uid=%s, gid=%s", user.Username, user.Uid, user.Gid)

		if err := c.WriteToFile(resourceFile); err != nil {
			return changed, files, message, fmt.Errorf("unable to write to file %s: %w", resourceFile, err)
		}
	}

	files = append(files, resourceFile)
	message = fmt.Sprintf("Update key %q from file %q", c.spec.Key, c.spec.File)

	return changed, files, message, err

}

func (c *CSV) singleTargetQuery(rootNode *dasel.Node) (changed bool, err error) {
	queryResult, err := rootNode.Query(c.spec.Key)

	if err != nil {
		// Catch error message returned by Dasel, if it couldn't find the node
		// This is approach is not very robust
		// https://github.com/TomWright/dasel/blob/master/node_query.go#L58

		if strings.HasPrefix(err.Error(), "could not find value:") {
			logrus.Debugln(err)
			err = fmt.Errorf("could not find value for query %q from file %q",
				c.spec.Key,
				c.spec.File)
			return changed, err
		}

		return changed, err
	}

	if queryResult == nil {
		err = fmt.Errorf("could not find value for query %q from file %q",
			c.spec.Key,
			c.spec.File)
		return changed, err
	}

	if queryResult.String() == c.spec.Value {
		logrus.Infof("%s Key %q, from file %q, already set to %q, nothing else need to do",
			result.SUCCESS,
			c.spec.Key,
			c.spec.File,
			c.spec.Value)
		return changed, nil
	}

	changed = true

	logrus.Infof("%s Key %q, from file %q, will be updated from  %q to %q",
		result.ATTENTION,
		c.spec.Key,
		c.spec.File,
		queryResult.String(),
		c.spec.Value)

	err = rootNode.Put(c.spec.Key, c.spec.Value)
	if err != nil {
		return changed, err
	}
	return changed, err
}

func (c *CSV) multipleTargetQuery(rootNode *dasel.Node) (changed bool, err error) {
	queryResults, err := rootNode.QueryMultiple(c.spec.Key)

	if err != nil {
		// Catch error message returned by Dasel, if it couldn't find the node
		// This is approach is not very robust
		// https://github.com/TomWright/dasel/blob/master/node_query.go#L58

		if strings.HasPrefix(err.Error(), "could not find multiple value:") {
			logrus.Debugln(err)
			err = fmt.Errorf("could not find multiple value for query %q from file %q",
				c.spec.Key,
				c.spec.File)
			return changed, err
		}

		return changed, err
	}

	if queryResults == nil {
		err = fmt.Errorf("could not find multiple value for query %q from file %q",
			c.spec.Key,
			c.spec.File)
		return changed, err
	}

	for i := range queryResults {

		queryResult := queryResults[i]

		if queryResult.String() == c.spec.Value {
			logrus.Infof("%s Key %q, from file %q, already set to %q, nothing else need to do",
				result.SUCCESS,
				c.spec.Key,
				c.spec.File,
				c.spec.Value)
			continue
		}

		logrus.Infof("%s Key %q, from file %q, will be updated from  %q to %q",
			result.ATTENTION,
			c.spec.Key,
			c.spec.File,
			queryResult.String(),
			c.spec.Value)

		changed = true

	}

	if !changed {
		logrus.Infof("%s Key(s) %q, from file %q, already set to %q, nothing else need to do",
			result.SUCCESS,
			c.spec.Key,
			c.spec.File,
			c.spec.Value)
		return changed, nil
	}

	err = rootNode.PutMultiple(c.spec.Key, c.spec.Value)
	if err != nil {
		return changed, err
	}
	return changed, err
}

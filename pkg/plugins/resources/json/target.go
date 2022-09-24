package json

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/tomwright/dasel"
	"github.com/tomwright/dasel/storage"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (j *Json) Target(source string, dryRun bool) (changed bool, err error) {

	changed, _, _, err = j.TargetFromSCM(source, nil, dryRun)
	if err != nil {
		return changed, err
	}

	return changed, err
}

// TargetFromSCM updates a scm repository based on the modified yaml file.
func (j *Json) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (changed bool, files []string, message string, err error) {

	if strings.HasPrefix(j.spec.File, "https://") ||
		strings.HasPrefix(j.spec.File, "http://") {
		return false, files, message, fmt.Errorf("URL scheme is not supported for Json target: %q", j.spec.File)
	}

	if scm != nil {
		j.spec.File = joinPathWithWorkingDirectoryPath(j.spec.File, scm.GetDirectory())
	}

	// Test at runtime if a file exist
	if !j.contentRetriever.FileExists(j.spec.File) {
		return false, files, message, fmt.Errorf("the Json file %q does not exist", j.spec.File)
	}

	if len(j.spec.Value) == 0 {
		j.spec.Value = source
	}

	resourceFile := ""
	if scm != nil {
		resourceFile = filepath.Join(scm.GetDirectory(), j.spec.File)
	} else {
		resourceFile = j.spec.File
	}

	if err := j.Read(); err != nil {
		return false, []string{}, "", err
	}

	// Override value from source if not yet defined
	if len(j.spec.Value) == 0 {
		j.spec.Value = source
	}

	var data interface{}

	err = json.Unmarshal([]byte(j.currentContent), &data)

	if err != nil {
		return false, []string{}, "", err
	}

	rootNode := dasel.New(data)

	if rootNode == nil {
		return changed, files, message, ErrDaselFailedParsingJSONByteFormat
	}

	queryResult, err := rootNode.Query(j.spec.Key)
	if err != nil {
		return changed, files, message, err
	}

	if queryResult.String() == j.spec.Value {
		logrus.Infof("%s Key %q, from file %q, already set to %q, nothing else need to do",
			result.SUCCESS,
			j.spec.Key,
			j.spec.File,
			j.spec.Value)
		return changed, files, message, nil
	}

	err = rootNode.Put(j.spec.Key, j.spec.Value)
	if err != nil {
		return changed, files, message, err
	}

	changed = true

	logrus.Infof("%s Key %q, from file %q, updated from  %q to %q",
		result.ATTENTION,
		j.spec.Key,
		j.spec.File,
		queryResult.String(),
		j.spec.Value)

	if !dryRun {

		fileInfo, err := os.Stat(resourceFile)
		if err != nil {
			return changed, files, message, fmt.Errorf("[%s] unable to get file info: %w", j.spec.File, err)
		}

		logrus.Debugf("fileInfo for %s mode=%s", resourceFile, fileInfo.Mode().String())

		user, err := user.Current()
		if err != nil {
			logrus.Errorf("unable to get user info: %s", err)
		}

		logrus.Debugf("user: username=%s, uid=%s, gid=%s", user.Username, user.Uid, user.Gid)

		newFile, err := os.Create(resourceFile)
		if err != nil {
			return changed, files, message, fmt.Errorf("unable to write to file %s: %w", resourceFile, err)
		}

		defer newFile.Close()

		err = rootNode.Write(
			newFile,
			"json",
			[]storage.ReadWriteOption{
				{
					Key:   storage.OptionIndent,
					Value: "  ",
				},
				{
					Key:   storage.OptionPrettyPrint,
					Value: true,
				},
			},
		)

		if err != nil {
			return changed, files, message, fmt.Errorf("unable to write to file %s: %w", resourceFile, err)
		}

	}

	files = append(files, resourceFile)
	message = fmt.Sprintf("Update key %q from file %q", j.spec.Key, j.spec.File)

	return changed, files, message, err

}

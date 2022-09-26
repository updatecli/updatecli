package toml

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml"
	"github.com/sirupsen/logrus"
	"github.com/tomwright/dasel"
	"github.com/tomwright/dasel/storage"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (t *Toml) Target(source string, dryRun bool) (changed bool, err error) {

	changed, _, _, err = t.TargetFromSCM(source, nil, dryRun)
	if err != nil {
		return changed, err
	}

	return changed, err
}

// TargetFromSCM updates a scm repository based on the modified yaml file.
func (t *Toml) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (changed bool, files []string, message string, err error) {

	if strings.HasPrefix(t.spec.File, "https://") ||
		strings.HasPrefix(t.spec.File, "http://") {
		return false, files, message, fmt.Errorf("URL scheme is not supported for Toml target: %q", t.spec.File)
	}

	if scm != nil {
		t.spec.File = joinPathWithWorkingDirectoryPath(t.spec.File, scm.GetDirectory())
	}

	// Test at runtime if a file exist
	if !t.contentRetriever.FileExists(t.spec.File) {
		return false, files, message, fmt.Errorf("the Toml file %q does not exist", t.spec.File)
	}

	if len(t.spec.Value) == 0 {
		t.spec.Value = source
	}

	resourceFile := ""
	if scm != nil {
		resourceFile = filepath.Join(scm.GetDirectory(), t.spec.File)
	} else {
		resourceFile = t.spec.File
	}

	if err := t.Read(); err != nil {
		return false, []string{}, "", err
	}

	// Override value from source if not yet defined
	if len(t.spec.Value) == 0 {
		t.spec.Value = source
	}

	var data interface{}

	err = toml.Unmarshal([]byte(t.currentContent), &data)

	if err != nil {
		return false, []string{}, "", err
	}

	rootNode := dasel.New(data)

	if rootNode == nil {
		return changed, files, message, ErrDaselFailedParsingTOMLByteFormat
	}

	queryResult, err := rootNode.Query(t.spec.Key)
	if err != nil {
		// Catch error message returned by Dasel, if it couldn't find the node
		// This is approach is not very robust
		// https://github.com/TomWright/dasel/blob/master/node_query.go#L58

		if strings.HasPrefix(err.Error(), "could not find value:") {
			logrus.Debugln(err)
			err = fmt.Errorf("could not find value for query %q from file %q",
				t.spec.Key,
				t.spec.File)
			return changed, files, message, err
		}

		return changed, files, message, err
	}

	if queryResult.String() == t.spec.Value {
		logrus.Infof("%s Key %q, from file %q, already set to %q, nothing else need to do",
			result.SUCCESS,
			t.spec.Key,
			t.spec.File,
			t.spec.Value)
		return changed, files, message, nil
	}

	err = rootNode.Put(t.spec.Key, t.spec.Value)
	if err != nil {
		return changed, files, message, err
	}

	changed = true

	logrus.Infof("%s Key %q, from file %q, updated from  %q to %q",
		result.ATTENTION,
		t.spec.Key,
		t.spec.File,
		queryResult.String(),
		t.spec.Value)

	if !dryRun {

		fileInfo, err := os.Stat(resourceFile)
		if err != nil {
			return changed, files, message, fmt.Errorf("[%s] unable to get file info: %w", t.spec.File, err)
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
			"toml",
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
	message = fmt.Sprintf("Update key %q from file %q", t.spec.Key, t.spec.File)

	return changed, files, message, err

}

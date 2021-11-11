package yaml

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/scm"
	"github.com/updatecli/updatecli/pkg/core/text"
	"gopkg.in/yaml.v3"
)

// Target updates a scm repository based on the modified yaml file.
func (y *Yaml) Target(source string, dryRun bool) (changed bool, err error) {
	y.Spec.Value = source

	// Test if target reference is an URL. In that case we don't know how to update.
	if text.IsURL(y.Spec.File) {
		return false, fmt.Errorf("unsupported filename prefix")
	}

	changed = false

	if err := y.Read(); err != nil {
		return false, err
	}
	data := y.CurrentContent

	out := yaml.Node{}

	err = yaml.Unmarshal([]byte(data), &out)

	if err != nil {
		return changed, fmt.Errorf("cannot unmarshal data: %v", err)
	}

	valueFound, oldVersion, _ := replace(&out, strings.Split(y.Spec.Key, "."), y.Spec.Value, 1)

	if valueFound {
		if oldVersion == y.Spec.Value {
			logrus.Infof("\u2714 Key '%s', from file '%v', already set to %s, nothing else need to be done",
				y.Spec.Key,
				y.Spec.File,
				y.Spec.Value)
			return changed, nil
		}

		changed = true
		logrus.Infof("\u2714 Key '%s', from file '%v', was updated from '%s' to '%s'",
			y.Spec.Key,
			y.Spec.File,
			oldVersion,
			y.Spec.Value)
	} else {
		logrus.Infof("\u2717 cannot find key '%s' from file '%s'", y.Spec.Key, y.Spec.File)
		return changed, nil
	}

	if !dryRun {
		fileInfo, err := os.Stat(y.Spec.File)
		if err != nil {
			logrus.Errorf("unable to get file info: %s", err)
		}

		logrus.Debugf("fileInfo for %s mode=%s", y.Spec.File, fileInfo.Mode().String())

		user, err := user.Current()
		if err != nil {
			logrus.Errorf("unable to get user info: %s", err)
		}

		logrus.Debugf("user: username=%s, uid=%s, gid=%s", user.Username, user.Uid, user.Gid)

		newFile, err := os.Create(y.Spec.File)
		if err != nil {
			return changed, fmt.Errorf("unable to write to file %s: %v", y.Spec.File, err)
		}

		defer newFile.Close()

		encoder := yaml.NewEncoder(newFile)
		defer encoder.Close()
		encoder.SetIndent(yamlIdent)
		err = encoder.Encode(&out)

		if err != nil {
			return changed, fmt.Errorf("something went wrong while encoding %v", err)
		}
	}

	return changed, nil
}

// TargetFromSCM updates a scm repository based on the modified yaml file.
func (y *Yaml) TargetFromSCM(source string, scm scm.Scm, dryRun bool) (changed bool, files []string, message string, err error) {
	if text.IsURL(y.Spec.File) {
		return false, files, message, fmt.Errorf("unsupported filename prefix")
	}

	y.Spec.Value = source

	changed = false

	if err := y.Read(); err != nil {
		return false, files, message, err
	}
	data := y.CurrentContent

	out := yaml.Node{}

	err = yaml.Unmarshal([]byte(data), &out)

	if err != nil {
		return changed, files, message, fmt.Errorf("cannot unmarshal data: %v", err)
	}

	valueFound, oldVersion, _ := replace(&out, strings.Split(y.Spec.Key, "."), y.Spec.Value, 1)

	if valueFound {
		if oldVersion == y.Spec.Value {
			logrus.Infof("\u2714 Key '%s', from file '%v', already set to %s, nothing else need to be done",
				y.Spec.Key,
				y.Spec.File,
				y.Spec.Value)
			return changed, files, message, nil
		}
		changed = true
		logrus.Infof("\u2714 Key '%s', from file '%v', was updated from '%s' to '%s'",
			y.Spec.Key,
			y.Spec.File,
			oldVersion,
			y.Spec.Value)

	} else {
		logrus.Infof("\u2717 cannot find key '%s' from file '%s'", y.Spec.Key, y.Spec.File)
		return changed, files, message, nil
	}

	if !dryRun {

		newFile, err := os.Create(filepath.Join(scm.GetDirectory(), y.Spec.File))
		defer newFile.Close()

		if err != nil {
			return changed, files, message, nil
		}

		encoder := yaml.NewEncoder(newFile)
		defer encoder.Close()
		encoder.SetIndent(yamlIdent)
		err = encoder.Encode(&out)

		if err != nil {
			return changed, files, message, err
		}
	}

	files = append(files, y.Spec.File)

	message = fmt.Sprintf("Update key %q from file %q", y.Spec.Key, y.Spec.File)

	return changed, files, message, nil
}

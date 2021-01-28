package yaml

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/olblak/updateCli/pkg/core/scm"
	"github.com/olblak/updateCli/pkg/plugins/file"
	"gopkg.in/yaml.v3"
)

// Target updates a scm repository based on the modified yaml file.
func (y *Yaml) Target(source string, dryRun bool) (changed bool, err error) {
	y.Value = source

	if len(y.Path) > 0 {
		logrus.Warnf("Key 'Path' is obsolete and now directly defined from file")
	}

	// Test if target reference a file with a prefix like https:// or file://
	// In that case we don't know how to update those files.
	if file.HasPrefix(y.File, []string{"https://", "http://", "file://"}) {
		return false, fmt.Errorf("unsupported filename prefix")
	}

	changed = false

	data, err := file.Read(y.File, "")
	if err != nil {
		return changed, err
	}

	out := yaml.Node{}

	err = yaml.Unmarshal(data, &out)

	if err != nil {
		return changed, fmt.Errorf("cannot unmarshal data: %v", err)
	}

	valueFound, oldVersion, _ := replace(&out, strings.Split(y.Key, "."), y.Value, 1)

	if valueFound {
		if oldVersion == y.Value {
			logrus.Infof("\u2714 Key '%s', from file '%v', already set to %s, nothing else need to be done",
				y.Key,
				y.File,
				y.Value)
			return changed, nil
		}

		changed = true
		logrus.Infof("\u2714 Key '%s', from file '%v', was updated from '%s' to '%s'",
			y.Key,
			y.File,
			oldVersion,
			y.Value)
	} else {
		logrus.Infof("\u2717 cannot find key '%s' from file '%s'", y.Key, y.File)
		return changed, nil
	}

	if !dryRun {
		fileInfo, err := os.Stat(y.File)
		if err != nil {
			logrus.Errorf("unable to get file info: %s", err)
		}

		var uid int
		var gid int
		if stat, ok := fileInfo.Sys().(*syscall.Stat_t); ok {
			uid = int(stat.Uid)
			gid = int(stat.Gid)
		} else {
			// we are not in linux, this won't work anyway in windows,
			// but maybe you want to log warnings
			uid = -1
			gid = -1
		}

		logrus.Debugf("fileInfo for %s mode=%s, uid=%d, gid=%d", y.File, fileInfo.Mode().String(), uid, gid)

		user, err := user.Current()
		if err != nil {
			logrus.Errorf("unable to get user info: %s", err)
		}

		logrus.Debugf("user: username=%s, uid=%s, gid=%s", user.Username, user.Uid, user.Gid)

		newFile, err := os.Create(y.File)
		if err != nil {
			return changed, fmt.Errorf("unable to write to file %s: %v", y.File, err)
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

	if len(y.Path) > 0 {
		logrus.Warnf("WARNING: Key 'Path' is obsolete and now directly retrieve from File")
	}

	// Test if target reference a file with a prefix like https:// or file://
	// In that case we don't know how to update those files.
	if file.HasPrefix(y.File, []string{"https://", "http://", "file://"}) {
		return false, files, message, fmt.Errorf("unsupported filename prefix")
	}

	y.Value = source

	changed = false

	data, err := file.Read(y.File, scm.GetDirectory())
	if err != nil {
		return changed, files, message, err
	}

	out := yaml.Node{}

	err = yaml.Unmarshal(data, &out)

	if err != nil {
		return changed, files, message, fmt.Errorf("cannot unmarshal data: %v", err)
	}

	valueFound, oldVersion, _ := replace(&out, strings.Split(y.Key, "."), y.Value, 1)

	if valueFound {
		if oldVersion == y.Value {
			logrus.Infof("\u2714 Key '%s', from file '%v', already set to %s, nothing else need to be done",
				y.Key,
				y.File,
				y.Value)
			return changed, files, message, nil
		}
		changed = true
		logrus.Infof("\u2714 Key '%s', from file '%v', was updated from '%s' to '%s'",
			y.Key,
			y.File,
			oldVersion,
			y.Value)

	} else {
		logrus.Infof("\u2717 cannot find key '%s' from file '%s'", y.Key, y.File)
		return changed, files, message, nil
	}

	if !dryRun {

		newFile, err := os.Create(filepath.Join(scm.GetDirectory(), y.File))
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

	files = append(files, y.File)
	message = fmt.Sprintf("[updatecli] Key '%s', from file '%v', was updated from %s to '%s'",
		y.Key,
		y.File,
		oldVersion,
		y.Value)

	return changed, files, message, nil
}

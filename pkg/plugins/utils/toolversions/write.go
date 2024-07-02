package toolversions

import (
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func (f *FileContent) Write() error {
	fileInfo, err := os.Stat(f.FilePath)
	if err != nil {
		return fmt.Errorf("[%s] unable to get file info: %w", f.FilePath, err)
	}

	logrus.Debugf("fileInfo for %s mode=%s", f.FilePath, fileInfo.Mode().String())

	user, err := user.Current()
	if err != nil {
		logrus.Errorf("unable to get user info: %s", err)
	}

	logrus.Debugf("user: username=%s, uid=%s, gid=%s", user.Username, user.Uid, user.Gid)

	fs := afero.NewOsFs()
	newFile, err := fs.Create(f.FilePath)
	if err != nil {
		return fmt.Errorf("unable to write to file %s: %w", f.FilePath, err)
	}

	defer newFile.Close()
	err = writeToolVersions(fs, newFile, f.Entries)
	if err != nil {
		return fmt.Errorf("unable to write to file %s: %w", f.FilePath, err)
	}

	return nil
}

// writeToolVersions creates or overwrites the .tool-versions file with the provided entries.
func writeToolVersions(fs afero.Fs, newFile afero.File, entries []Entry) error {
	for _, entry := range entries {
		line := fmt.Sprintf("%s %s\n", entry.Key, strings.TrimSpace(entry.Value))
		if _, err := newFile.WriteString(line); err != nil {
			return err
		}
	}
	return nil
}

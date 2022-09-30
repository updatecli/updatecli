package dasel

import (
	"fmt"
	"os"
	"os/user"

	"github.com/sirupsen/logrus"
	"github.com/tomwright/dasel/storage"
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

	newFile, err := os.Create(f.FilePath)
	if err != nil {
		return fmt.Errorf("unable to write to file %s: %w", f.FilePath, err)
	}

	defer newFile.Close()

	switch f.DataType {
	case "json":
		err = f.DaselNode.Write(
			newFile,
			f.DataType,
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
			return fmt.Errorf("unable to write to file %s: %w", f.FilePath, err)
		}

	case "toml":
		err = f.DaselNode.Write(
			newFile,
			f.DataType,
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
			return fmt.Errorf("unable to write to file %s: %w", f.FilePath, err)
		}
	default:
		return fmt.Errorf("data type %q no supported", f.DataType)
	}

	return nil
}

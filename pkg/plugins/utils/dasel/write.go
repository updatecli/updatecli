package dasel

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"

	"github.com/BurntSushi/toml"
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
	case TYPEJSON:
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
				{
					Key:   storage.OptionEscapeHTML,
					Value: false,
				},
			},
		)
		if err != nil {
			return fmt.Errorf("unable to write to file %s: %w", f.FilePath, err)
		}

	case TYPETOML:
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

// WriteV3 serializes the native parsed data (DaselV3Data) modified by the dasel v3
// engine back to disk. It marshals with the standard encoders so the output format
// matches the v1/v2 engines (2-space indentation, HTML escaping disabled for json)
// rather than the dasel v3 native writer, keeping diffs stable across engines.
func (f *FileContent) WriteV3() error {
	if f.DaselV3Data == nil {
		return ErrEmptyDaselNode
	}

	newFile, err := os.Create(f.FilePath)
	if err != nil {
		return fmt.Errorf("unable to write to file %s: %w", f.FilePath, err)
	}

	defer newFile.Close()

	switch f.DataType {
	case TYPEJSON:
		encoder := json.NewEncoder(newFile)
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)

		if err := encoder.Encode(f.DaselV3Data); err != nil {
			return fmt.Errorf("unable to write to file %s: %w", f.FilePath, err)
		}

	case TYPETOML:
		encoder := toml.NewEncoder(newFile)
		encoder.Indent = "  "

		if err := encoder.Encode(f.DaselV3Data); err != nil {
			return fmt.Errorf("unable to write to file %s: %w", f.FilePath, err)
		}

	default:
		return fmt.Errorf("data type %q not supported by the dasel v3 engine", f.DataType)
	}

	return nil
}

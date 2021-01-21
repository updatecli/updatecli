package file

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Source return a file content
func (f *File) Source(workingDir string) (string, error) {

	var data []byte
	var err error

	if strings.HasPrefix(f.File, "https://") ||
		strings.HasPrefix(f.File, "http://") {
		data, err = ReadFromURL(f.File)

		if err != nil {
			return "", err
		}

	} else if strings.HasPrefix(f.File, "file://") {
		f.File = strings.TrimPrefix(f.File, "file://")

		data, err = ReadFromFile(filepath.Join(workingDir, f.File))

		if err != nil {
			return "", err
		}
	} else {
		data, err = ReadFromFile(filepath.Join(workingDir, f.File))

		if err != nil {
			return "", err
		}

	}

	fmt.Printf("\u2714 Content:\n%v\n\n found from file %v \n",
		Show(string(data)),
		f.File)

	return string(data), nil
}

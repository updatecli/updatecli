package file

import (
	"github.com/sirupsen/logrus"
)

// Source return a file content
func (f *File) Source(workingDir string) (string, error) {

	data, err := Read(f.File, workingDir)
	if err != nil {
		return "", err
	}

	logrus.Infof("\u2714 Content:\n%v\n\n found from file %v",
		Show(string(data)),
		f.File)

	return string(data), nil
}

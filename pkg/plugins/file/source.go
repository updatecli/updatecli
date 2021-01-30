package file

import (
	"strings"

	"github.com/sirupsen/logrus"
)

// Source return a file content
func (f *File) Source(workingDir string) (string, error) {

	data, err := Read(f.File, workingDir)
	if err != nil {
		return "", err
	}

	if len(f.Content) == 0 {
		f.Content = string(data)
	}

	if len(f.Line) > 0 {
		for _, line := range strings.Split(f.Content, "\n") {
			if strings.Contains(line, f.Line) {
				f.Content = line
				break
			}
		}
	}

	logrus.Infof("\u2714 Content:\n%v\n\n found from file %v", f.Content, f.File)

	return f.Content, nil
}

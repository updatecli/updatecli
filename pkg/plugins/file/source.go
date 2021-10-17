package file

import (
	"github.com/sirupsen/logrus"
)

// Source return a file content
func (f *File) Source(workingDir string) (string, error) {
	if err := f.Read(); err != nil {
		return "", err
	}

	logrus.Infof("\u2714 Content: found from file %q:\n%v", f.spec.File, f.CurrentContent)

	return f.CurrentContent, nil
}

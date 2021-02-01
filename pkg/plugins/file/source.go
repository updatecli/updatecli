package file

import (
	"fmt"

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

	if len(f.Line.Excludes) > 0 {
		f.Content, err = f.Line.ContainsExcluded(f.Content)

		if err != nil {
			return "", err
		}

	}

	if len(f.Line.HasIncludes) > 0 {
		if ok, err := f.Line.HasIncluded(f.Content); err != nil || !ok {
			if err != nil {
				return "", err
			}

			if !ok {
				return "", fmt.Errorf(ErrLineNotFound)
			}

		}
	}

	if len(f.Line.Includes) > 0 {
		f.Content, err = f.Line.ContainsIncluded(f.Content)

		if err != nil {
			return "", err
		}

	}

	logrus.Infof("\u2714 Content:\n%v\n\n found from file %v", f.Content, f.File)

	return f.Content, nil
}

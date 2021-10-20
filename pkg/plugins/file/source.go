package file

import (
	"fmt"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/text"
)

// Source return a file content
func (f *File) Source(workingDir string) (string, error) {

	data, err := text.ReadAll(filepath.Join(workingDir, f.File))
	if err != nil {
		return "", err
	}

	if len(f.Content) == 0 {
		f.Content = data
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

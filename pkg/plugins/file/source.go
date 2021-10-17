package file

import (
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/text"
)

// Source return a file content
func (f *File) Source(workingDir string) (string, error) {

	data, err := text.ReadAll(filepath.Join(workingDir, f.spec.File))
	if err != nil {
		return "", err
	}

	if len(f.spec.Content) == 0 {
		f.spec.Content = data
	}

	if len(f.spec.Line.Excludes) > 0 {
		f.spec.Content, err = f.spec.Line.ContainsExcluded(f.spec.Content)

		if err != nil {
			return "", err
		}

	}

	if len(f.spec.Line.HasIncludes) > 0 {
		if ok, err := f.spec.Line.HasIncluded(f.spec.Content); err != nil || !ok {
			if err != nil {
				return "", err
			}

			if !ok {
				return "", &ErrLineNotFound{}
			}

		}
	}

	if len(f.spec.Line.Includes) > 0 {
		f.spec.Content, err = f.spec.Line.ContainsIncluded(f.spec.Content)

		if err != nil {
			return "", err
		}

	}

	logrus.Infof("\u2714 Content:\n%v\n\n found from file %v", f.spec.Content, f.spec.File)

	return f.spec.Content, nil
}

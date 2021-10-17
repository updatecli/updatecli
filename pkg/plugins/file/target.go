package file

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/scm"
	"github.com/updatecli/updatecli/pkg/core/text"
)

// Target creates or updates a file located locally.
// The default content is the value retrieved from source
func (f *File) Target(source string, dryRun bool) (changed bool, err error) {

	if len(f.spec.Content) == 0 {
		f.spec.Content = source
	}

	// Test if target reference a file with a prefix like https:// or file://
	// In that case we don't know how to update those files.
	if text.IsURL(f.spec.File) {
		return false, fmt.Errorf("Unsupported filename prefix")
	}

	changed = false

	data, err := text.ReadAll(f.spec.File)
	if err != nil {
		return false, err
	}

	content := data

	if len(f.spec.Line.Excludes) > 0 {
		f.spec.Content, err = f.spec.Line.ContainsExcluded(f.spec.Content)

		if err != nil {
			return false, err
		}

	}

	if len(f.spec.Line.HasIncludes) > 0 {
		if ok, err := f.spec.Line.HasIncluded(f.spec.Content); err != nil || !ok {
			if err != nil {
				return false, err
			}

			if !ok {
				return false, &ErrLineNotFound{}
			}

		}

	}

	if len(f.spec.Line.Includes) > 0 {
		f.spec.Content, err = f.spec.Line.ContainsIncluded(f.spec.Content)

		if err != nil {
			return false, err
		}
	}

	if strings.Compare(f.spec.Content, content) == 0 {
		logrus.Infof("\u2714 Content from file '%v' already up to date", f.spec.File)
		return false, nil
	}

	changed = true
	logrus.Infof("\u2714 File content for '%v', updated. \n%s",
		f.spec.File, text.Diff(content, f.spec.Content))

	if !dryRun {

		err := text.WriteToFile(f.spec.Content, f.spec.File)
		if err != nil {
			return false, err
		}
	}

	return changed, nil
}

// TargetFromSCM creates or updates a file from a source control management system.
// The default content is the value retrieved from source
func (f *File) TargetFromSCM(source string, scm scm.Scm, dryRun bool) (changed bool, files []string, message string, err error) {

	if len(f.spec.Content) == 0 {
		f.spec.Content = source
	}

	// Test if target reference a file with a prefix like https:// or file://
	// In that case we don't know how to update those files.
	if text.IsURL(f.spec.File) {
		return false, files, message, fmt.Errorf("unsupported filename prefix")
	}

	changed = false

	data, err := text.ReadAll(filepath.Join(scm.GetDirectory(), f.spec.File))
	if err != nil {
		return changed, files, message, err
	}

	if strings.Compare(f.spec.Content, data) != 0 {
		changed = true
		logrus.Infof("\u2714 File content for '%v', updated. \n%s",
			f.spec.File, text.Diff(data, f.spec.Content))

	} else {
		logrus.Infof("\u2714 Content from file '%v' already up to date", f.spec.File)
	}

	if !dryRun {

		err := text.WriteToFile(f.spec.Content, filepath.Join(scm.GetDirectory(), f.spec.File))
		if err != nil {
			return false, files, message, err
		}
	}

	files = append(files, f.spec.File)
	message = fmt.Sprintf("Update %q content\n", f.spec.File)

	return changed, files, message, nil
}

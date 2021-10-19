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

	if len(f.Content) == 0 {
		f.Content = source
	}

	// Test if target reference a file with a prefix like https:// or file://
	// In that case we don't know how to update those files.
	if text.IsURL(f.File) {
		return false, fmt.Errorf("Unsupported filename prefix")
	}

	changed = false

	data, err := text.ReadAll(f.File)
	if err != nil {
		return false, err
	}

	content := data

	if len(f.Line.Excludes) > 0 {
		f.Content, err = f.Line.ContainsExcluded(f.Content)

		if err != nil {
			return false, err
		}

	}

	if len(f.Line.HasIncludes) > 0 {
		if ok, err := f.Line.HasIncluded(f.Content); err != nil || !ok {
			if err != nil {
				return false, err
			}

			if !ok {
				return false, fmt.Errorf(ErrLineNotFound)
			}

		}

	}

	if len(f.Line.Includes) > 0 {
		f.Content, err = f.Line.ContainsIncluded(f.Content)

		if err != nil {
			return false, err
		}
	}

	if strings.Compare(f.Content, content) == 0 {
		logrus.Infof("\u2714 Content from file '%v' already up to date", f.File)
		return false, nil
	}

	changed = true
	logrus.Infof("\u2714 File content for '%v', updated. \n%s",
		f.File, text.Diff(content, f.Content))

	if !dryRun {

		err := text.WriteToFile(f.Content, f.File)
		if err != nil {
			return false, err
		}
	}

	return changed, nil
}

// TargetFromSCM creates or updates a file from a source control management system.
// The default content is the value retrieved from source
func (f *File) TargetFromSCM(source string, scm scm.Scm, dryRun bool) (changed bool, files []string, message string, err error) {

	if len(f.Content) == 0 {
		f.Content = source
	}

	// Test if target reference a file with a prefix like https:// or file://
	// In that case we don't know how to update those files.
	if text.IsURL(f.File) {
		return false, files, message, fmt.Errorf("unsupported filename prefix")
	}

	changed = false

	data, err := text.ReadAll(filepath.Join(f.File, scm.GetDirectory()))
	if err != nil {
		return changed, files, message, err
	}

	if strings.Compare(f.Content, data) != 0 {
		changed = true
		logrus.Infof("\u2714 File content for '%v', updated. \n%s",
			f.File, text.Diff(data, f.Content))

	} else {
		logrus.Infof("\u2714 Content from file '%v' already up to date", f.File)
	}

	if !dryRun {

		err := text.WriteToFile(f.Content, filepath.Join(scm.GetDirectory(), f.File))
		if err != nil {
			return false, files, message, err
		}
	}

	files = append(files, f.File)
	message = fmt.Sprintf("Update %q content\n", f.File)

	return changed, files, message, nil
}

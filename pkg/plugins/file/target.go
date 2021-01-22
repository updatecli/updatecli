package file

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/olblak/updateCli/pkg/core/scm"
)

// Target creates or updates a file located locally.
// The default content is the value retrieved from source
func (f *File) Target(source string, dryRun bool) (changed bool, err error) {
	data := []byte{}

	if len(f.Content) == 0 {
		f.Content = source
	}

	workingDir := ""

	changed = false

	if strings.HasPrefix(f.File, "https://") ||
		strings.HasPrefix(f.File, "http://") {

		if err != nil {
			return false, fmt.Errorf("Target don't support file using HTTP/HTTPS url")
		}

	} else if strings.HasPrefix(f.File, "file://") {
		f.File = strings.TrimPrefix(f.File, "file://")

		data, err = ReadFromFile(filepath.Join(workingDir, f.File))

		if err != nil {
			return false, err
		}
	} else {
		data, err = ReadFromFile(filepath.Join(workingDir, f.File))

		if err != nil {
			return false, err
		}

	}

	if strings.Compare(f.Content, string(data)) != 0 {
		changed = true
		fmt.Printf("\u2714 Content from file '%v', updated. \n%s\n",
			f.File, Diff(f.Content, string(data)))

	} else {
		fmt.Printf("\u2714 Content from file '%v' already up to date\n", f.File)
	}

	if !dryRun {

		err := WriteToFile(f.Content, f.File)
		if err != nil {
			return false, err
		}
	}

	return changed, nil
}

// TargetFromSCM creates or updates a file from a source control management system.
// The default content is the value retrieved from source
func (f *File) TargetFromSCM(source string, scm scm.Scm, dryRun bool) (changed bool, files []string, message string, err error) {
	data := []byte{}

	if len(f.Content) == 0 {
		f.Content = source
	}

	workingDir := ""

	changed = false

	if strings.HasPrefix(f.File, "https://") ||
		strings.HasPrefix(f.File, "http://") {

		if err != nil {
			return false, files, message, fmt.Errorf("Target don't support file using HTTP url")
		}

	} else if strings.HasPrefix(f.File, "file://") {
		f.File = strings.TrimPrefix(f.File, "file://")

		data, err = ReadFromFile(filepath.Join(workingDir, f.File))

		if err != nil {
			return false, files, message, err
		}
	} else {
		data, err = ReadFromFile(filepath.Join(workingDir, f.File))

		if err != nil {
			return false, files, message, err
		}

	}

	if strings.Compare(f.Content, string(data)) != 0 {
		changed = true
		fmt.Printf("\u2714 Content from file '%v', updated. \n%s\n",
			f.File, Diff(f.Content, string(data)))

	} else {
		fmt.Printf("\u2714 Content from file '%v' already up to date\n", f.File)
	}

	if !dryRun {

		err := WriteToFile(f.Content, f.File)
		if err != nil {
			return false, files, message, err
		}
	}

	files = append(files, f.File)
	message = fmt.Sprintf("[updatecli] Content for file '%v' updated\n", f.File)

	return changed, files, message, nil
}

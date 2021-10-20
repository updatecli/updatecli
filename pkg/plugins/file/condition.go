package file

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/scm"
	"github.com/updatecli/updatecli/pkg/core/text"
)

// Condition test if a file content match the content provided via configuration.
// If the configuration doesn't specify a value then it fall back to the source output
func (f *File) Condition(source string) (bool, error) {

	// If f.Content is not provided then we use the value returned from source
	if len(f.Content) == 0 {
		f.Content = source
	}

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
		logrus.Infof("\u2714 Content from file '%v' is correct'", f.File)
		return true, nil
	}

	logrus.Infof("\u2717 Wrong content from file '%v'. \n%s",
		f.File, text.Diff(f.Content, content))

	return false, nil
}

// ConditionFromSCM test if a file content from SCM match the content provided via configuration.
// If the configuration doesn't specify a value then it fall back to the source output
func (f *File) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {

	if len(f.Content) > 0 {
		logrus.Infof("Key content defined from updatecli configuration")
	} else {
		f.Content = source
	}

	data, err := text.ReadAll(filepath.Join(scm.GetDirectory(), f.File))
	if err != nil {
		return false, err
	}

	if strings.Compare(f.Content, data) == 0 {
		logrus.Infof("\u2714 Content from file '%v' is correct'", f.File)
		return true, nil
	}

	logrus.Infof("\u2717 Wrong content from file '%v'. \n%s",
		f.File, text.Diff(f.Content, data))

	return false, nil
}

package file

import (
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/scm"
	"github.com/updatecli/updatecli/pkg/core/text"
)

// Condition test if a file content match the content provided via configuration.
// If the configuration doesn't specify a value then it fall back to the source output
func (f *File) Condition(source string) (bool, error) {

	// If f.spec.Content is not provided then we use the value returned from source
	if len(f.spec.Content) == 0 {
		f.spec.Content = source
	}

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
		logrus.Infof("\u2714 Content from file '%v' is correct'", f.spec.File)
		return true, nil
	}

	logrus.Infof("\u2717 Wrong content from file '%v'. \n%s",
		f.spec.File, text.Diff(f.spec.Content, content))

	return false, nil
}

// ConditionFromSCM test if a file content from SCM match the content provided via configuration.
// If the configuration doesn't specify a value then it fall back to the source output
func (f *File) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {

	if len(f.spec.Content) > 0 {
		logrus.Infof("Key content defined from updatecli configuration")
	} else {
		f.spec.Content = source
	}

	data, err := text.ReadAll(filepath.Join(scm.GetDirectory(), f.spec.File))
	if err != nil {
		return false, err
	}

	if strings.Compare(f.spec.Content, data) == 0 {
		logrus.Infof("\u2714 Content from file '%v' is correct'", f.spec.File)
		return true, nil
	}

	logrus.Infof("\u2717 Wrong content from file '%v'. \n%s",
		f.spec.File, text.Diff(f.spec.Content, data))

	return false, nil
}

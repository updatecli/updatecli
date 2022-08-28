package xml

import (
	"fmt"
	"path/filepath"

	"github.com/beevik/etree"
	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (x *XML) Target(source string, dryRun bool) (changed bool, err error) {

	changed, _, _, err = x.TargetFromSCM(source, nil, dryRun)
	if err != nil {
		return changed, err
	}

	return changed, err
}

// TargetFromSCM updates a scm repository based on the modified yaml file.
func (x *XML) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (changed bool, files []string, message string, err error) {
	if len(x.spec.Value) == 0 {
		x.spec.Value = source
	}

	resourceFile := ""
	if scm != nil {
		resourceFile = filepath.Join(scm.GetDirectory(), x.spec.File)
	} else {
		resourceFile = x.spec.File
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromFile(resourceFile); err != nil {
		return false, []string{}, "", err
	}

	elem := doc.FindElement(x.spec.Path)

	if elem == nil {
		return false, []string{}, "", fmt.Errorf("%s nothing found at path %q from file %q",
			result.FAILURE,
			x.spec.Path,
			x.spec.File)
	}

	if elem.Text() == x.spec.Value {
		logrus.Infof("%s Path '%s', from file '%v', already set to %s, nothing else need to be done",
			result.SUCCESS,
			x.spec.Path,
			x.spec.File,
			x.spec.Value)
		return false, []string{}, "", nil
	}
	logrus.Infof("%s Key '%s', from file '%v', was updated from '%s' to '%s'",
		result.ATTENTION,
		x.spec.Path,
		x.spec.File,
		elem.Text(),
		x.spec.Value)

	if !dryRun {
		elem.SetText(x.spec.Value)

		if err := doc.WriteToFile(resourceFile); err != nil {
			return false, []string{}, "", err
		}
	}

	files = append(files, resourceFile)
	message = fmt.Sprintf("Update key %q from file %q", x.spec.Path, x.spec.File)

	return true, files, message, nil

}

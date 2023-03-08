package xml

import (
	"fmt"

	"github.com/beevik/etree"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (x *XML) Condition(source string) (bool, error) {
	return x.ConditionFromSCM(source, nil)
}

func (x *XML) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {

	resourceFile := x.spec.File
	if scm != nil {
		resourceFile = joinPathWithWorkingDirectoryPath(x.spec.File, scm.GetDirectory())
	}

	value := source
	if len(x.spec.Value) > 0 {
		value = x.spec.Value
	}

	// Test at runtime if a file exist
	if !x.contentRetriever.FileExists(resourceFile) {
		return false, fmt.Errorf("file %q does not exist", resourceFile)
	}

	if err := x.Read(resourceFile); err != nil {
		return false, err
	}

	doc := etree.NewDocument()

	if err := doc.ReadFromString(x.currentContent); err != nil {
		return false, err
	}

	elem := doc.FindElement(x.spec.Path)

	if elem == nil {
		logrus.Infof("%s nothing found in path %q from file %q",
			result.FAILURE,
			x.spec.Path,
			resourceFile)

		return false, nil
	}

	if value == elem.Text() {
		logrus.Infof("%s Path %q, from file %q, is correctly set to %s",
			result.SUCCESS,
			x.spec.Path,
			resourceFile,
			value)
		return true, nil
	}

	logrus.Infof("%s Path %q, from file %q, is incorrectly set to %q and should be %q",
		result.ATTENTION,
		x.spec.Path,
		resourceFile,
		elem.Text(),
		value)

	return false, nil
}

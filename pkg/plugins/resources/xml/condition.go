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

	if scm != nil {
		x.spec.File = joinPathWithWorkingDirectoryPath(x.spec.File, scm.GetDirectory())
	}

	// Test at runtime if a file exist
	if !x.contentRetriever.FileExists(x.spec.File) {
		return false, fmt.Errorf("the XML file %q does not exist", x.spec.File)
	}

	if err := x.Read(); err != nil {
		return false, err
	}

	doc := etree.NewDocument()

	if err := doc.ReadFromString(x.currentContent); err != nil {
		return false, err
	}

	// Override value from source if not yet defined
	if len(x.spec.Value) == 0 {
		x.spec.Value = source
	}

	elem := doc.FindElement(x.spec.Path)

	if elem == nil {
		logrus.Infof("%s nothing found in path '%s' from file '%s'",
			result.FAILURE,
			x.spec.Path,
			x.spec.File)

		return false, nil
	}

	if x.spec.Value == elem.Text() {
		logrus.Infof("%s Path '%s', from file '%v', is correctly set to %s'",
			result.SUCCESS,
			x.spec.Path,
			x.spec.File,
			x.spec.Value)
		return true, nil
	}

	logrus.Infof("%s Path '%s', from file '%v', is incorrectly set to %s and should be %s'",
		result.ATTENTION,
		x.spec.Path,
		x.spec.File,
		elem.Text(),
		x.spec.Value)

	return false, nil
}

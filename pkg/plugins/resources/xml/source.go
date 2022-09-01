package xml

import (
	"path/filepath"

	"github.com/beevik/etree"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (x *XML) Source(workingDir string) (string, error) {

	doc := etree.NewDocument()
	if err := doc.ReadFromFile(filepath.Join(workingDir, x.spec.File)); err != nil {
		return "", err
	}

	elem := doc.FindElement(x.spec.Path)

	if elem == nil {
		logrus.Infof("%s cannot find value for path '%s' from file '%s'",
			result.FAILURE,
			x.spec.Path,
			x.spec.File)

		return "", nil
	}

	queryResult := elem.Text()

	logrus.Infof("%s Value %q found at path %q in the xml file %q",
		result.SUCCESS, queryResult, x.spec.Path, x.spec.File)

	return queryResult, nil
}

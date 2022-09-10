package xml

import (
	"fmt"

	"github.com/beevik/etree"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (x *XML) Source(workingDir string) (string, error) {

	// To merge File path with current working dire, unless file is an http url
	x.spec.File = joinPathWithWorkingDirectoryPath(x.spec.File, workingDir)

	// Test at runtime if a file exist
	if !x.contentRetriever.FileExists(x.spec.File) {
		return "", fmt.Errorf("the XML file %q does not exist", x.spec.File)
	}

	if err := x.Read(); err != nil {
		return "", err
	}

	doc := etree.NewDocument()

	if err := doc.ReadFromString(x.currentContent); err != nil {
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

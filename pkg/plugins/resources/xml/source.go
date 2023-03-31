package xml

import (
	"errors"
	"fmt"
	"os"

	"github.com/beevik/etree"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (x *XML) Source(workingDir string) (string, error) {

	// By the default workingdir is set to the current working directory
	// it would be better to have it empty by default but it must be changed in the
	// souce core codebase.
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		return "", errors.New("fail getting current working directory")
	}

	resourceFile := x.spec.File
	// To merge File path with current working dire, unless file is an http url
	if workingDir != currentWorkingDirectory {
		resourceFile = joinPathWithWorkingDirectoryPath(x.spec.File, workingDir)
	}

	// Test at runtime if a file exist
	if !x.contentRetriever.FileExists(resourceFile) {
		return "", fmt.Errorf("file %q does not exist", resourceFile)
	}

	if err := x.Read(resourceFile); err != nil {
		return "", err
	}

	doc := etree.NewDocument()

	if err := doc.ReadFromString(x.currentContent); err != nil {
		return "", err
	}

	elem := doc.FindElement(x.spec.Path)

	if elem == nil {
		logrus.Infof("%s cannot find value for path %q from file %q",
			result.FAILURE,
			x.spec.Path,
			resourceFile)

		return "", nil
	}

	queryResult := elem.Text()

	logrus.Infof("%s Value %q found at path %q in the xml file %q",
		result.SUCCESS,
		queryResult,
		x.spec.Path,
		resourceFile)

	return queryResult, nil
}

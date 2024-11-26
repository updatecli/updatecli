package xml

import (
	"errors"
	"fmt"
	"os"

	"github.com/beevik/etree"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns a value from a xml file
func (x *XML) Source(workingDir string, resultSource *result.Source) error {

	// By the default workingdir is set to the current working directory
	// it would be better to have it empty by default but it must be changed in the
	// source core codebase.
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		return errors.New("fail getting current working directory")
	}

	resourceFile := x.spec.File
	// To merge File path with current working dire, unless file is an http url
	if workingDir != currentWorkingDirectory {
		resourceFile = joinPathWithWorkingDirectoryPath(x.spec.File, workingDir)
	}

	// Test at runtime if a file exist
	if !x.contentRetriever.FileExists(resourceFile) {
		return fmt.Errorf("file %q does not exist", resourceFile)
	}

	if err := x.Read(resourceFile); err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	doc := etree.NewDocument()

	if err := doc.ReadFromString(x.currentContent); err != nil {
		return fmt.Errorf("loading document: %w", err)
	}

	elem := doc.FindElement(x.spec.Path)

	if elem == nil {
		return fmt.Errorf("cannot find value for path %q from file %q",
			x.spec.Path,
			resourceFile,
		)
	}

	queryResult := elem.Text()

	resultSource.Result = result.SUCCESS
	resultSource.Information = []result.SourceInformation{{
		Key:   resultSource.ID,
		Value: queryResult,
	}}
	resultSource.Description = fmt.Sprintf("value %q found at path %q in the xml file %q",
		queryResult,
		x.spec.Path,
		resourceFile)

	return nil
}

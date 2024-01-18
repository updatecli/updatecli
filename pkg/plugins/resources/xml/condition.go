package xml

import (
	"fmt"

	"github.com/beevik/etree"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Condition checks that a specific xml path contains the correct value at the specified path
func (x *XML) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {

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
		return false, "", fmt.Errorf("file %q does not exist", resourceFile)
	}

	if err := x.Read(resourceFile); err != nil {
		return false, "", err
	}

	doc := etree.NewDocument()

	if err := doc.ReadFromString(x.currentContent); err != nil {
		return false, "", err
	}

	elem := doc.FindElement(x.spec.Path)

	if elem == nil {
		return false, fmt.Sprintf("nothing found in path %q from file %q",
			x.spec.Path,
			resourceFile,
		), nil
	}

	if value == elem.Text() {
		return true, fmt.Sprintf("Path %q, from file %q, is correctly set to %s",
			x.spec.Path,
			resourceFile,
			value), nil
	}

	return false, fmt.Sprintf("Path %q, from file %q, is incorrectly set to %q and should be %q",
		x.spec.Path,
		resourceFile,
		elem.Text(),
		value,
	), nil
}

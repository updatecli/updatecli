package xml

import (
	"fmt"

	"github.com/beevik/etree"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition checks that a specific xml path contains the correct value at the specified path
func (x *XML) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {

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
		return fmt.Errorf("file %q does not exist", resourceFile)
	}

	if err := x.Read(resourceFile); err != nil {
		return err
	}

	doc := etree.NewDocument()

	if err := doc.ReadFromString(x.currentContent); err != nil {
		return err
	}

	elem := doc.FindElement(x.spec.Path)

	if elem == nil {
		resultCondition.Description = fmt.Sprintf("nothing found in path %q from file %q",
			x.spec.Path,
			resourceFile,
		)

		resultCondition.Pass = false
		resultCondition.Result = result.FAILURE

		return nil
	}

	if value == elem.Text() {
		resultCondition.Description = fmt.Sprintf("%s Path %q, from file %q, is correctly set to %s",
			result.SUCCESS,
			x.spec.Path,
			resourceFile,
			value)

		resultCondition.Pass = true
		resultCondition.Result = result.SUCCESS

		return nil
	}

	resultCondition.Description = fmt.Sprintf("Path %q, from file %q, is incorrectly set to %q and should be %q",
		x.spec.Path,
		resourceFile,
		elem.Text(),
		value,
	)
	resultCondition.Pass = false
	resultCondition.Result = result.FAILURE

	return nil
}

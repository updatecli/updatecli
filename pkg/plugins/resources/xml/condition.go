package xml

import (
	"context"
	"fmt"
	"strings"

	"github.com/beevik/etree"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

// Condition checks that a specific xml path contains the correct value at the specified path
func (x *XML) Condition(_ context.Context, source string, scm scm.ScmHandler, pathResolver pathresolver.Resolver) (pass bool, message string, err error) {

	resourceFile, err := pathResolver.Resolve(x.spec.File)
	if err != nil {
		return false, "", fmt.Errorf("invalid file path %q: %w", x.spec.File, err)
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

	// The whitespace around a value, such as the indentation of an element spanning several lines, is ignored
	if strings.TrimSpace(value) == elemText(elem) {
		return true, fmt.Sprintf("Path %q, from file %q, is correctly set to %s",
			x.spec.Path,
			resourceFile,
			value), nil
	}

	return false, fmt.Sprintf("Path %q, from file %q, is incorrectly set to %q and should be %q",
		x.spec.Path,
		resourceFile,
		elemText(elem),
		value,
	), nil
}

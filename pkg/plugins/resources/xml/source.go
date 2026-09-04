package xml

import (
	"context"
	"fmt"

	"github.com/beevik/etree"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
)

// Source returns a value from a xml file
func (x *XML) Source(_ context.Context, resolver utils.Resolver, resultSource *result.Source) error {

	resourceFile, err := resolver.Resolve(x.spec.File)
	if err != nil {
		return fmt.Errorf("invalid file path %q: %w", x.spec.File, err)
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
	resultSource.Information = queryResult
	resultSource.Description = fmt.Sprintf("value %q found at path %q in the xml file %q",
		queryResult,
		x.spec.Path,
		resourceFile)

	return nil
}

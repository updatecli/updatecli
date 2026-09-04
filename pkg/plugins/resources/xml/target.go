package xml

import (
	"context"
	"fmt"
	"strings"

	"github.com/beevik/etree"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
)

// Target updates a scm repository based on the modified yaml file.
func (x *XML) Target(_ context.Context, source string, scm scm.ScmHandler, resolver utils.Resolver, dryRun bool, resultTarget *result.Target) (err error) {

	if strings.HasPrefix(x.spec.File, "https://") ||
		strings.HasPrefix(x.spec.File, "http://") {
		return fmt.Errorf("URL scheme is not supported for XML target: %q", x.spec.File)
	}

	value := source
	if x.spec.Value != "" {
		value = x.spec.Value
	}

	resultTarget.NewInformation = value

	resourceFile, err := resolver.Resolve(x.spec.File)
	if err != nil {
		return fmt.Errorf("invalid file path %q: %w", x.spec.File, err)
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
		return fmt.Errorf("nothing found at path %q from file %q", x.spec.Path, resourceFile)
	}

	resultTarget.Information = elem.Text()
	resultTarget.NewInformation = value

	if elem.Text() == value {
		resultTarget.Result = result.SUCCESS
		resultTarget.Description = fmt.Sprintf("path %q already set to %q in file %q",
			x.spec.Path,
			value,
			resourceFile)
		return nil
	}
	resultTarget.Result = result.ATTENTION
	resultTarget.Changed = true

	resultTarget.Description = fmt.Sprintf("path %q updated from %q to %q in file %q",
		x.spec.Path,
		elem.Text(),
		value,
		resourceFile)

	if !dryRun {
		elem.SetText(value)

		if err := doc.WriteToFile(resourceFile); err != nil {
			return err
		}
	}

	resultTarget.Files = append(resultTarget.Files, x.spec.File)

	return nil
}

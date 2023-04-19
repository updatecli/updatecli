package xml

import (
	"fmt"
	"strings"

	"github.com/beevik/etree"
	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Target updates a scm repository based on the modified yaml file.
func (x *XML) Target(source string, scm scm.ScmHandler, dryRun bool) (changed bool, files []string, message string, err error) {

	if strings.HasPrefix(x.spec.File, "https://") ||
		strings.HasPrefix(x.spec.File, "http://") {
		return false, files, message, fmt.Errorf("URL scheme is not supported for XML target: %q", x.spec.File)
	}

	value := source
	if len(x.spec.Value) > 0 {
		value = x.spec.Value
	}

	resourceFile := x.spec.File
	if scm != nil {
		resourceFile = joinPathWithWorkingDirectoryPath(x.spec.File, scm.GetDirectory())
	}

	// Test at runtime if a file exist
	if !x.contentRetriever.FileExists(resourceFile) {
		return false, files, message, fmt.Errorf("file %q does not exist", resourceFile)
	}

	if err := x.Read(resourceFile); err != nil {
		return false, []string{}, "", err
	}

	doc := etree.NewDocument()

	if err := doc.ReadFromString(x.currentContent); err != nil {
		return false, []string{}, "", err
	}

	elem := doc.FindElement(x.spec.Path)
	if elem == nil {
		return false, []string{}, "", fmt.Errorf("%s nothing found at path %q from file %q",
			result.FAILURE,
			x.spec.Path,
			resourceFile)
	}

	if elem.Text() == value {
		logrus.Infof("%s Path %q, from file %q, already set to %q, nothing else to do",
			result.SUCCESS,
			x.spec.Path,
			resourceFile,
			value)
		return false, []string{}, "", nil
	}
	logrus.Infof("%s Key %q, from file '%v', was updated from %q to %q",
		result.ATTENTION,
		x.spec.Path,
		resourceFile,
		elem.Text(),
		value)

	if !dryRun {
		elem.SetText(value)

		if err := doc.WriteToFile(resourceFile); err != nil {
			return false, []string{}, "", err
		}
	}

	files = append(files, x.spec.File)
	message = fmt.Sprintf("Update key %q from file %q",
		x.spec.Path, x.spec.File)

	return true, files, message, nil
}

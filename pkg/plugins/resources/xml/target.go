package xml

import (
	"fmt"
	"strings"

	"github.com/beevik/etree"
	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/result"
)

func (x *XML) Target(source string, dryRun bool) (changed bool, err error) {
	if strings.HasPrefix(x.spec.File, "https://") ||
		strings.HasPrefix(x.spec.File, "http://") {
		return false, fmt.Errorf("URL scheme is not supported for XML target: %q", x.spec.File)
	}

	// Test at runtime if a file exist
	if !x.contentRetriever.FileExists(x.spec.File) {
		return false, fmt.Errorf("the XML file %q does not exist", x.spec.File)
	}

	if len(x.spec.Value) == 0 {
		x.spec.Value = source
	}

	resourceFile := x.spec.File

	if err := x.Read(); err != nil {
		return false, err
	}

	doc := etree.NewDocument()

	if err := doc.ReadFromString(x.currentContent); err != nil {
		return false, err
	}

	elem := doc.FindElement(x.spec.Path)

	if elem == nil {
		return false, fmt.Errorf("%s nothing found at path %q from file %q",
			result.FAILURE,
			x.spec.Path,
			x.spec.File)
	}

	if elem.Text() == x.spec.Value {
		logrus.Infof("%s Path '%s', from file '%v', already set to %s, nothing else need to be done",
			result.SUCCESS,
			x.spec.Path,
			x.spec.File,
			x.spec.Value)
		return false, nil
	}
	logrus.Infof("%s Key '%s', from file '%v', was updated from '%s' to '%s'",
		result.ATTENTION,
		x.spec.Path,
		x.spec.File,
		elem.Text(),
		x.spec.Value)

	if !dryRun {
		elem.SetText(x.spec.Value)

		if err := doc.WriteToFile(resourceFile); err != nil {
			return false, err
		}
	}

	return true, nil
}

package toolversions

import (
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Target updates a scm repository based on the modified .tool-versions file.
func (t *ToolVersions) Target(source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {

	rootDir := ""
	if scm != nil {
		rootDir = scm.GetDirectory()
	}

	for i := range t.contents {
		filename := t.contents[i].FilePath

		// Target doesn't support updating files on remote http location
		if strings.HasPrefix(filename, "https://") ||
			strings.HasPrefix(filename, "http://") {
			return fmt.Errorf("URL scheme is not supported for toolversions target: %q", t.spec.File)
		}

		if err := t.contents[i].Read(rootDir); err != nil {
			return fmt.Errorf("file %q does not exist", t.contents[i].FilePath)
		}

		if len(t.spec.Value) == 0 {
			t.spec.Value = source
		}
		resultTarget.NewInformation = t.spec.Value

		resourceFile := t.contents[i].FilePath

		// Override value from source if not yet defined
		if len(t.spec.Value) == 0 {
			t.spec.Value = source
		}

		resultTarget.Description = fmt.Sprintf("%s\nkey %q, from file %q, is correctly set to %q",
			resultTarget.Description,
			t.spec.Key,
			t.contents[i].FilePath,
			t.spec.Value)

		if dryRun {
			continue
		}

		err := t.contents[i].Put(t.spec.Key, t.spec.Value)
		if err != nil {
			return err
		}

		err = t.contents[i].Write()
		if err != nil {
			return err
		}

		resultTarget.Files = append(resultTarget.Files, resourceFile)
	}

	resultTarget.Description = strings.TrimPrefix(resultTarget.Description, "\n")

	return nil
}

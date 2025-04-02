package toolversions

import (
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

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
			return fmt.Errorf("file %q does not exist", filename)
		}

		if len(t.spec.Value) == 0 {
			t.spec.Value = source
		}
		resultTarget.NewInformation = t.spec.Value

		// Override value from source if not yet defined
		if len(t.spec.Value) == 0 {
			t.spec.Value = source
		}

		queryResult, _ := t.contents[i].Get(t.spec.Key)

		if !t.spec.CreateMissingKey && queryResult == "" {
			return fmt.Errorf("key %q does not exist. Use createMissingKey if you want to create the key", t.spec.Key)
		}

		changedFile := false
		switch queryResult == resultTarget.NewInformation {
		case true:
			resultTarget.Description = fmt.Sprintf("%s\nkey %q, from file %q, is correctly set to %q",
				resultTarget.Description,
				t.spec.Key,
				filename,
				t.spec.Value)

		case false:
			changedFile = true
			resultTarget.Information = queryResult
			resultTarget.Result = result.ATTENTION
			resultTarget.Changed = true
			resultTarget.Description = fmt.Sprintf("%s\nkey %q, from file %q, is incorrectly set to %q and should be %q",
				resultTarget.Description,
				t.spec.Key,
				filename,
				queryResult,
				t.spec.Value)
		}

		if !changedFile || dryRun {
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

		resultTarget.Files = append(resultTarget.Files, filename)
	}

	resultTarget.Description = strings.TrimPrefix(resultTarget.Description, "\n")

	return nil
}

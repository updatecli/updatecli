package toolversions

import (
	"errors"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
)

func (t *ToolVersions) Source(workingDir string, resultSource *result.Source) error {

	if len(t.contents) > 1 {
		return errors.New("source only supports one file")
	}

	content := t.contents[0]

	if err := content.Read(workingDir); err != nil {
		return fmt.Errorf("reading .tool-versions file: %w", err)
	}

	key := t.spec.Key
	value, err := content.Get(key)
	if err != nil {
		return err
	}

	resultSource.Information = []result.SourceInformation{{
		Key:   resultSource.ID,
		Value: value,
	}}
	resultSource.Result = result.SUCCESS
	resultSource.Description = fmt.Sprintf("value %q, found in file %q, for key %q'",
		value,
		content.FilePath,
		key)

	return nil
}

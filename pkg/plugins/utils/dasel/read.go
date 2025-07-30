package dasel

import (
	"encoding/json"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/tomwright/dasel"
)

// Read reads the content of a file after runtime validation
func (f *FileContent) Read(rootDir string) error {

	f.FilePath = JoinPathWithWorkingDirectoryPath(f.FilePath, rootDir)

	if !f.ContentRetriever.FileExists(f.FilePath) {
		return fmt.Errorf("file %q does not exist", f.FilePath)
	}

	textContent, err := f.ContentRetriever.ReadAll(
		f.FilePath)

	if err != nil {
		return fmt.Errorf("failed to read file %q: %w", f.FilePath, err)
	}

	var data any
	switch f.DataType {

	case "json":
		err = json.Unmarshal([]byte(textContent), &data)
		if err != nil {
			return fmt.Errorf("failed to unmarshal json content: %w", err)
		}

	case "toml":
		err := toml.Unmarshal([]byte(textContent), &data)

		if err != nil {
			return fmt.Errorf("failed to unmarshal toml content: %w", err)
		}

	default:
		return fmt.Errorf("%q datatype not support", f.DataType)
	}

	daselNode := dasel.New(data)
	f.DaselNode = daselNode

	f.DaselV2Node = data

	if f.DaselNode == nil || f.DaselV2Node == nil {
		return ErrDaselFailedParsingByteFormat
	}

	return nil
}

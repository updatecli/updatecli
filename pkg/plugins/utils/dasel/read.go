package dasel

import (
	"encoding/json"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/tomwright/dasel"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

// ResolvePath resolves FilePath against the path resolver.
//
// It is idempotent: the path written in the manifest is kept aside so that calling it
// twice cannot join the base directory twice.
func (f *FileContent) ResolvePath(pathResolver pathresolver.Resolver) error {
	if f.OriginalFilePath == "" {
		f.OriginalFilePath = f.FilePath
	}

	resolvedPath, err := pathResolver.Resolve(f.OriginalFilePath)
	if err != nil {
		return err
	}

	f.FilePath = resolvedPath

	return nil
}

// Read reads the content of a file after runtime validation
func (f *FileContent) Read(pathResolver pathresolver.Resolver) error {

	if err := f.ResolvePath(pathResolver); err != nil {
		return err
	}

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

	case TYPEJSON:
		err = json.Unmarshal([]byte(textContent), &data)
		if err != nil {
			return fmt.Errorf("failed to unmarshal json content: %w", err)
		}

	case TYPETOML:
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

	f.DaselV3Data = data

	if f.DaselNode == nil || f.DaselV2Node == nil {
		return ErrDaselFailedParsingByteFormat
	}

	return nil
}

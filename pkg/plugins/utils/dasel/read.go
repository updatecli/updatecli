package dasel

import (
	"encoding/json"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/tomwright/dasel"
)

// Read reads the content of a file after runtime validation
func (f *FileContent) Read(rootDir string) error {

	f.FilePath = joinPathWithWorkingDirectoryPath(f.FilePath, rootDir)

	if !f.ContentRetriever.FileExists(f.FilePath) {
		return fmt.Errorf("file %q does not exist", f.FilePath)
	}

	textContent, err := f.ContentRetriever.ReadAll(
		f.FilePath)

	if err != nil {
		return err
	}

	var data interface{}
	switch f.DataType {

	case "json":
		err = json.Unmarshal([]byte(textContent), &data)
		if err != nil {
			return err
		}

	case "toml":
		err := toml.Unmarshal([]byte(textContent), &data)

		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("%q datatype not support", f.DataType)
	}

	daselNode := dasel.New(data)

	if daselNode != nil {
		f.DaselNode = daselNode
		return nil
	}

	return ErrDaselFailedParsingByteFormat

}

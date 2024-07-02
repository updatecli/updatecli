package toolversions

/*
 Toolversions package provides an abstraction of the .tool-versions file.
*/

import (
	"errors"

	"github.com/updatecli/updatecli/pkg/core/text"
)

var (
	// ErrToolVersionsFailedParsingByteFormat is returned if tool-versions couldn't parse the byteData
	ErrToolVersionsFailedParsingByteFormat error = errors.New("failed to parse file")
)

type FileContent struct {
	// FilePath defines the fullpath filename
	FilePath string
	// ContentRetriever is an interface to manipulate raw files
	ContentRetriever text.TextRetriever
	// Entries contains the .tool-versions representation of the file
	Entries []Entry
}

// Entry represents a key-value pair in the .tool-versions file.
type Entry struct {
	Key   string
	Value string
}

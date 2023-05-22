package dasel

/*
 Dasel package provides an abstraction of https://github.com/TomWright/dasel/
 So it's easier to integrate it in Updatecli by various data file such as Json,Toml, or csv
*/

import (
	"errors"

	"github.com/tomwright/dasel"
	"github.com/updatecli/updatecli/pkg/core/text"
)

var (
	// ErrDaselFailedParsingByteFormat is returned if dasel couldn't parse the byteData
	ErrDaselFailedParsingByteFormat error = errors.New("failed to parse file")
	// ErrEmptyDaselNode is returned when we try to manipulate a null Dasel node
	ErrEmptyDaselNode error = errors.New("no Dasel data")
)

type FileContent struct {
	// DataType defines what type of Dasel file we have, accepted value ["json"]
	DataType string
	// FilePath defines the fullpath filename
	FilePath string
	// ContentRetriever is an interface to manipulate raw files
	ContentRetriever text.TextRetriever
	// DaselNode contains the dasel representation of the file
	DaselNode *dasel.Node
}

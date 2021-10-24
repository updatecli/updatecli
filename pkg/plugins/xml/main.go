package xml

import (
	"errors"
)

var (
	// ErrDaselFailedParsingXMLByteFormat is returned if dasel couldn't parse the byteData
	ErrDaselFailedParsingXMLByteFormat error = errors.New("fail to parse XML data")
)

// XML stores configuration about the file and the key value which needs to be updated.
type XML struct {
	File  string
	Key   string
	Value string
}

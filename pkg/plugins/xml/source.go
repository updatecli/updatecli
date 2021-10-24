package xml

import (
	"path/filepath"

	"github.com/clbanning/mxj/v2"
	"github.com/tomwright/dasel"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func (x *XML) Source(workingDir string) (string, error) {

	strData, err := text.ReadAll(filepath.Join(workingDir, x.File))
	if err != nil {
		return "", err
	}

	data, err := mxj.NewMapXml([]byte(strData))

	if err != nil {
		return "", err
	}

	rootNode := dasel.New(data)

	if rootNode == nil {
		return "", ErrDaselFailedParsingXMLByteFormat
	}

	result, err := rootNode.Query(x.Key)
	if err != nil {
		return "", err
	}

	return result.String(), err
}

package xml

import (
	"path/filepath"

	"github.com/clbanning/mxj/v2"
	"github.com/sirupsen/logrus"
	"github.com/tomwright/dasel"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func (x *XML) Source(workingDir string) (string, error) {

	strData, err := text.ReadAll(filepath.Join(workingDir, x.spec.File))
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

	queryResult, err := rootNode.Query(x.spec.Key)

	if err != nil {
		return "", err
	}

	logrus.Infof("%s Value %q found for key %q in the xml file %q",
		result.SUCCESS, queryResult.String(), x.spec.Key, x.spec.File)

	return queryResult.String(), err
}

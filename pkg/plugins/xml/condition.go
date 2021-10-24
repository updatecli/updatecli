package xml

import (
	"path/filepath"

	"github.com/clbanning/mxj/v2"
	"github.com/sirupsen/logrus"
	"github.com/tomwright/dasel"
	"github.com/updatecli/updatecli/pkg/core/scm"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func (x *XML) Condition(source string) (bool, error) {

	strData, err := text.ReadAll(x.File)
	if err != nil {
		return false, err
	}

	if len(x.Value) == 0 {
		x.Value = source
	}

	data, err := mxj.NewMapXml([]byte(strData))

	if err != nil {
		return false, err
	}

	rootNode := dasel.New(data)

	if rootNode == nil {
		return false, ErrDaselFailedParsingXMLByteFormat
	}

	result, err := rootNode.Query(x.Key)
	if err != nil {
		return false, err
	}

	if result.String() == x.Value {
		logrus.Infof("\u2714 Key '%s', from file '%v', is correctly set to %s'",
			x.Key,
			x.File,
			x.Value)
		return true, nil
	}

	logrus.Infof("\u2717 Key '%s', from file '%v', is incorrectly set to %s and should be %s'",
		x.Key,
		x.File,
		result.String(),
		x.Value)

	return false, err
}

func (x *XML) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {
	strData, err := text.ReadAll(filepath.Join(scm.GetDirectory(), x.File))
	if err != nil {
		return false, err
	}

	if len(x.Value) == 0 {
		x.Value = source
	}

	data, err := mxj.NewMapXml([]byte(strData))

	if err != nil {
		return false, err
	}

	rootNode := dasel.New(data)

	if rootNode == nil {
		return false, ErrDaselFailedParsingXMLByteFormat
	}

	result, err := rootNode.Query(x.Key)
	if err != nil {
		return false, err
	}

	if result.String() == x.Value {
		logrus.Infof("\u2714 Key '%s', from file '%v', is correctly set to %s'",
			x.Key,
			x.File,
			x.Value)
		return true, nil
	} else {
		logrus.Infof("\u2717 Key '%s', from file '%v', is incorrectly set to %s and should be %s'",
			x.Key,
			x.File,
			result.String(),
			x.Value)
	}

	return true, err
}

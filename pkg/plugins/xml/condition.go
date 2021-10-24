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

	strData, err := text.ReadAll(x.spec.File)
	if err != nil {
		return false, err
	}

	if len(x.spec.Value) == 0 {
		x.spec.Value = source
	}

	data, err := mxj.NewMapXml([]byte(strData))

	if err != nil {
		return false, err
	}

	rootNode := dasel.New(data)

	if rootNode == nil {
		return false, ErrDaselFailedParsingXMLByteFormat
	}

	result, err := rootNode.Query(x.spec.Key)
	if err != nil {
		return false, err
	}

	if result.String() == x.spec.Value {
		logrus.Infof("\u2714 Key '%s', from file '%v', is correctly set to %s'",
			x.spec.Key,
			x.spec.File,
			x.spec.Value)
		return true, nil
	}

	logrus.Infof("\u2717 Key '%s', from file '%v', is incorrectly set to %s and should be %s'",
		x.spec.Key,
		x.spec.File,
		result.String(),
		x.spec.Value)

	return false, err
}

func (x *XML) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {
	strData, err := text.ReadAll(filepath.Join(scm.GetDirectory(), x.spec.File))
	if err != nil {
		return false, err
	}

	if len(x.spec.Value) == 0 {
		x.spec.Value = source
	}

	data, err := mxj.NewMapXml([]byte(strData))

	if err != nil {
		return false, err
	}

	rootNode := dasel.New(data)

	if rootNode == nil {
		return false, ErrDaselFailedParsingXMLByteFormat
	}

	result, err := rootNode.Query(x.spec.Key)
	if err != nil {
		return false, err
	}

	if result.String() == x.spec.Value {
		logrus.Infof("\u2714 Key '%s', from file '%v', is correctly set to %s'",
			x.spec.Key,
			x.spec.File,
			x.spec.Value)
		return true, nil
	} else {
		logrus.Infof("\u2717 Key '%s', from file '%v', is incorrectly set to %s and should be %s'",
			x.spec.Key,
			x.spec.File,
			result.String(),
			x.spec.Value)
	}

	return true, err
}

package xml

import (
	"path/filepath"

	"github.com/clbanning/mxj/v2"
	"github.com/sirupsen/logrus"
	"github.com/tomwright/dasel"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/scm"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func (x *XML) Condition(source string) (bool, error) {
	return x.ConditionFromSCM(source, nil)
}

func (x *XML) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {
	conditionFile := ""
	if scm != nil {
		conditionFile = filepath.Join(scm.GetDirectory(), x.spec.File)
	} else {
		conditionFile = x.spec.File
	}
	strData, err := text.ReadAll(conditionFile)
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

	queryResult, err := rootNode.Query(x.spec.Key)
	if err != nil {
		return false, err
	}

	if queryResult.String() == x.spec.Value {
		logrus.Infof("%s Key '%s', from file '%v', is correctly set to %s'",
			result.SUCCESS,
			x.spec.Key,
			x.spec.File,
			x.spec.Value)
		return true, nil
	} else {
		logrus.Infof("%s Key '%s', from file '%v', is incorrectly set to %s and should be %s'",
			result.ATTENTION,
			x.spec.Key,
			x.spec.File,
			queryResult.String(),
			x.spec.Value)
	}

	return false, err
}

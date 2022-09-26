package toml

import (
	"fmt"
	"strings"

	toml "github.com/pelletier/go-toml"

	"github.com/sirupsen/logrus"
	"github.com/tomwright/dasel"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (t *Toml) Condition(source string) (bool, error) {
	return t.ConditionFromSCM(source, nil)
}

func (t *Toml) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	if scm != nil {
		t.spec.File = joinPathWithWorkingDirectoryPath(t.spec.File, scm.GetDirectory())
	}

	// Test at runtime if a file exist
	if !t.contentRetriever.FileExists(t.spec.File) {
		return false, fmt.Errorf("the Toml file %q does not exist", t.spec.File)
	}

	if err := t.Read(); err != nil {
		return false, err
	}

	// Override value from source if not yet defined
	if len(t.spec.Value) == 0 {
		t.spec.Value = source
	}

	var data interface{}

	err := toml.Unmarshal([]byte(t.currentContent), &data)

	if err != nil {
		return false, err
	}

	rootNode := dasel.New(data)

	if rootNode == nil {
		return false, ErrDaselFailedParsingTOMLByteFormat
	}

	queryResult, err := rootNode.Query(t.spec.Key)
	if err != nil {
		// Catch error message returned by Dasel, if it couldn't find the node
		// This is approach is not very robust
		// https://github.com/TomWright/dasel/blob/master/node_query.go#L58

		if strings.HasPrefix(err.Error(), "could not find value:") {
			err = fmt.Errorf("%s could not find value for query %q from file %q",
				result.FAILURE,
				t.spec.Key,
				t.spec.File)
			return false, err
		}

		return false, err
	}

	if queryResult.String() == t.spec.Value {
		logrus.Infof("%s Key '%s', from file '%v', is correctly set to %s'",
			result.SUCCESS,
			t.spec.Key,
			t.spec.File,
			t.spec.Value)
		return true, nil
	} else {
		logrus.Infof("%s Key '%s', from file '%v', is incorrectly set to %s and should be %s'",
			result.ATTENTION,
			t.spec.Key,
			t.spec.File,
			queryResult.String(),
			t.spec.Value)
	}
	return false, nil
}

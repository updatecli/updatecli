package toml

import (
	"fmt"
	"strings"

	toml "github.com/pelletier/go-toml"
	"github.com/sirupsen/logrus"
	"github.com/tomwright/dasel"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (t *Toml) Source(workingDir string) (string, error) {

	// Test at runtime if a file exist
	if !t.contentRetriever.FileExists(t.spec.File) {
		return "", fmt.Errorf("the Toml file %q does not exist", t.spec.File)
	}

	if err := t.Read(); err != nil {
		return "", err
	}

	var data interface{}

	err := toml.Unmarshal([]byte(t.currentContent), &data)

	if err != nil {
		return "", err
	}

	rootNode := dasel.New(data)

	if rootNode == nil {
		return "", ErrDaselFailedParsingTOMLByteFormat
	}

	queryResult, err := rootNode.Query(t.spec.Key)
	if err != nil {
		// Catch error message returned by Dasel, if it couldn't find the node
		// This is approach is not very robust
		// https://github.com/TomWright/dasel/blob/master/node_query.go#L58

		if strings.HasPrefix(err.Error(), "could not find value:") {
			logrus.Debugln(err)
			err = fmt.Errorf("could not find value for query %q from file %q",
				t.spec.Key,
				t.spec.File)
			return "", err
		}
		return "", err
	}

	logrus.Infof("%s Value %q, found in file %q, for key %q'",
		result.SUCCESS,
		queryResult.String(),
		t.spec.File,
		t.spec.Key)

	return queryResult.String(), nil
}

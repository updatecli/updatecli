package yaml

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/text"

	"gopkg.in/yaml.v3"
)

// Source return the latest version
func (y *Yaml) Source(workingDir string) (string, error) {
	// By default workingDir is set to local directory

	if y.Value != "" {
		logrus.Warnf("Key 'Value' is not used by source YAML")
	}

	if len(y.Path) > 0 {
		logrus.Warnf("Key 'Path' is obsolete and now directly defined from file")
	}

	contentRetriever := &text.Text{}
	data, err := contentRetriever.ReadAll(filepath.Join(workingDir, y.File))
	if err != nil {
		return "", err
	}

	var out yaml.Node

	err = yaml.Unmarshal([]byte(data), &out)

	if err != nil {
		return "", fmt.Errorf("cannot unmarshal data: %v", err)
	}

	valueFound, value, _ := replace(&out, strings.Split(y.Key, "."), y.Value, 1)

	if valueFound {
		logrus.Infof("\u2714 Value '%v' found for key %v in the yaml file %v", value, y.Key, y.File)
		return value, nil
	}

	logrus.Infof("\u2717 cannot find key '%s' from file '%s'",
		y.Key,
		y.File)
	return "", nil

}

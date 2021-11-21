package yaml

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"

	"gopkg.in/yaml.v3"
)

// Source return the latest version
func (y *Yaml) Source(workingDir string) (string, error) {
	// By default workingDir is set to local directory

	if y.Spec.Value != "" {
		logrus.Warnf("Key 'Value' is not used by source YAML")
	}

	if err := y.Read(); err != nil {
		return "", err
	}

	data := y.CurrentContent

	var out yaml.Node

	err := yaml.Unmarshal([]byte(data), &out)

	if err != nil {
		return "", fmt.Errorf("cannot unmarshal data: %v", err)
	}

	valueFound, value, _ := replace(&out, strings.Split(y.Spec.Key, "."), y.Spec.Value, 1)

	if valueFound {
		logrus.Infof("%s Value '%v' found for key %v in the yaml file %v", result.SUCCESS, value, y.Spec.Key, y.Spec.File)
		return value, nil
	}

	logrus.Infof("%s cannot find key '%s' from file '%s'",
		result.FAILURE,
		y.Spec.Key,
		y.Spec.File)
	return "", nil

}

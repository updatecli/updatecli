package yaml

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"

	"gopkg.in/yaml.v3"
)

// Source return the latest version
func (y *Yaml) Source(workingDir string) (string, error) {
	// By default workingDir is set to local directory
	// Merge File path with current workingDir, unless File is an HTTP URL
	y.spec.File = joinPathWithWorkingDirectoryPath(y.spec.File, workingDir)

	// Test at runtime if a file exist
	if !y.contentRetriever.FileExists(y.spec.File) {
		return "", fmt.Errorf("the yaml file %q does not exist", y.spec.File)
	}

	if y.spec.Value != "" {
		logrus.Warnf("Key 'Value' is not used by source YAML")
	}

	if err := y.Read(); err != nil {
		return "", err
	}

	data := y.currentContent

	var out yaml.Node

	err := yaml.Unmarshal([]byte(data), &out)

	if err != nil {
		return "", fmt.Errorf("cannot unmarshal data: %v", err)
	}

	valueFound, value, _ := replace(&out, parseKey(y.spec.Key), y.spec.Value, 1)

	if valueFound {
		logrus.Infof("%s Value '%v' found for key %v in the yaml file %v", result.SUCCESS, value, y.spec.Key, y.spec.File)
		return value, nil
	}

	logrus.Infof("%s cannot find key '%s' from file '%s'",
		result.FAILURE,
		y.spec.Key,
		y.spec.File)
	return "", nil

}

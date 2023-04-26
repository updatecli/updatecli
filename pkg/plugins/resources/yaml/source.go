package yaml

import (
	"errors"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"

	"gopkg.in/yaml.v3"
)

// Source return the latest version
func (y *Yaml) Source(workingDir string) (string, error) {
	// By default workingDir is set to local directory
	var filePath string

	// By the default workingdir is set to the current working directory
	// it would be better to have it empty by default but it must be changed in the
	// source core codebase.
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		return "", errors.New("fail getting current working directory")
	}

	if len(y.files) > 1 {
		validationError := fmt.Errorf("validation error in sources of type 'yaml': the attributes `spec.files` can't contain more than one element for sources")
		logrus.Errorf(validationError.Error())
		return "", validationError
	}

	if y.spec.Value != "" {
		logrus.Warnf("Key 'Value' is not used by source YAML")
	}

	// loop over the only file
	for f := range y.files {
		filePath = f

		// Ideally currentWorkingDirectory should be empty
		if workingDir != currentWorkingDirectory {
			y.UpdateAbsoluteFilePath(workingDir)
		}
	}

	if err = y.Read(); err != nil {
		return "", err
	}

	fileContent := y.files[filePath].content
	originalFilePath := y.files[filePath].originalFilePath

	var out yaml.Node

	err = yaml.Unmarshal([]byte(fileContent), &out)
	if err != nil {
		return "", fmt.Errorf("cannot unmarshal content of file %s: %v", originalFilePath, err)
	}

	valueFound, value, _ := replace(&out, parseKey(y.spec.Key), y.spec.Value, 1)

	if valueFound {
		logrus.Infof("%s Value '%v' found for key %v in the yaml file %v",
			result.SUCCESS,
			value,
			y.spec.Key,
			originalFilePath,
		)
		return value, nil
	}

	logrus.Infof("%s cannot find key '%s' from file '%s'",
		result.FAILURE,
		y.spec.Key,
		originalFilePath)
	return "", nil

}

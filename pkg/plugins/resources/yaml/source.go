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
	var fileContent string
	var filePath string

	// By the default workingdir is set to the current working directory
	// it would be better to have it empty by default but it must be changed in the
	// souce core codebase.
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		return "", errors.New("fail getting current working directory")
	}

	if len(y.files) > 1 {
		validationError := fmt.Errorf("validation error in sources of type 'yaml': the attributes `spec.files` can't contain more than one element for conditions")
		logrus.Errorf(validationError.Error())
		return "", validationError
	}

	if y.spec.Value != "" {
		logrus.Warnf("Key 'Value' is not used by source YAML")
	}

	var errs []error
	// loop over the only file
	for theFilePath := range y.files {
		fileContent = y.files[theFilePath]
		filePath = theFilePath

		// Ideally currentWorkingDirectory should be empty
		if workingDir != currentWorkingDirectory {
			logrus.Debugf("current working directory set to %q", workingDir)
			filePath = joinPathWithWorkingDirectoryPath(filePath, workingDir)
		}

		// Test at runtime if a file exist
		if !y.contentRetriever.FileExists(filePath) {
			errs = append(errs, fmt.Errorf("the yaml file %q does not exist", filePath))
			continue
		}

		y.files[filePath], err = y.contentRetriever.ReadAll(filePath)
		if err != nil {
			errs = append(errs, fmt.Errorf("fail reading file %q, skipping", filePath))
			continue
		}
	}

	if len(errs) > 0 {
		for i := range errs {
			logrus.Errorf("\t * %s\n", errs[i])
		}
		return "", fmt.Errorf("fail reading yaml files (%d/%d)", len(errs), len(y.files))
	}

	var out yaml.Node

	err = yaml.Unmarshal([]byte(fileContent), &out)

	if err != nil {
		return "", fmt.Errorf("cannot unmarshal content of file %s: %v", filePath, err)
	}

	valueFound, value, _ := replace(&out, parseKey(y.spec.Key), y.spec.Value, 1)

	if valueFound {
		logrus.Infof("%s Value '%v' found for key %v in the yaml file %v", result.SUCCESS, value, y.spec.Key, filePath)
		return value, nil
	}

	logrus.Infof("%s cannot find key '%s' from file '%s'",
		result.FAILURE,
		y.spec.Key,
		filePath)
	return "", nil

}

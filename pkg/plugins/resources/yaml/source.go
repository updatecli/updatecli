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
func (y *Yaml) Source(workingDir string, resultSource *result.Source) error {
	// By default workingDir is set to local directory
	var filePath string

	// By the default workingdir is set to the current working directory
	// it would be better to have it empty by default but it must be changed in the
	// source core codebase.
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		return errors.New("fail getting current working directory")
	}

	if len(y.files) > 1 {
		validationError := fmt.Errorf("validation error in sources of type 'yaml': the attributes `spec.files` can't contain more than one element for sources")
		logrus.Errorf(validationError.Error())
		return validationError
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
		return fmt.Errorf("reading yaml file: %w", err)
	}

	fileContent := y.files[filePath].content
	originalFilePath := y.files[filePath].originalFilePath

	var out yaml.Node

	err = yaml.Unmarshal([]byte(fileContent), &out)
	if err != nil {
		return fmt.Errorf("unmarshalling content of file %s: %w", originalFilePath, err)
	}

	valueFound, value, _ := replace(&out, parseKey(y.spec.Key), y.spec.Value, 1)

	if valueFound {
		resultSource.Result = result.SUCCESS
		resultSource.Information = value
		resultSource.Description = fmt.Sprintf("value %q found for key %q in the yaml file %q",
			value,
			y.spec.Key,
			originalFilePath,
		)
		return nil
	}

	return fmt.Errorf("cannot find key %q from file %q",
		y.spec.Key,
		originalFilePath)

}

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

	var fileContent string
	var filePath string

	if len(y.files) > 1 {
		validationError := fmt.Errorf("Validation error in sources of type 'yaml': the attributes `spec.files` can't contain more than one element for conditions")
		logrus.Errorf(validationError.Error())
		return "", validationError
  }

  // Merge File path with current workingDir, unless File is an HTTP URL
	y.spec.File = joinPathWithWorkingDirectoryPath(y.spec.File, workingDir)

	// Test at runtime if a file exist
	if !y.contentRetriever.FileExists(y.spec.File) {
		return "", fmt.Errorf("the yaml file %q does not exist", y.spec.File)
	}

	if y.spec.Value != "" {
		logrus.Warnf("Key 'Value' is not used by source YAML")
	}

	// TODO: warn if the boolean 'KeyOnly' is set?

	if err := y.Read(); err != nil {
		return "", err
	}

	// loop over the only file
	for theFilePath := range y.files {
		fileContent = y.files[theFilePath]
		filePath = theFilePath
	}

	var out yaml.Node

	err := yaml.Unmarshal([]byte(fileContent), &out)

	if err != nil {
		return "", fmt.Errorf("cannot unmarshal content of file %s: %v", filePath, err)
	}

	valueFound, value, _ := replace(&out, strings.Split(y.spec.Key, "."), y.spec.Value, 1)

	if valueFound {
		logrus.Infof("%s Value '%v' found for key %v in the yaml file %v", result.SUCCESS, value, y.spec.Key, filePath)
		return value, nil
	}

	// TODO: return result.WARNING? Or an actual error?
	logrus.Infof("%s cannot find key '%s' from file '%s'",
		result.FAILURE,
		y.spec.Key,
		filePath)
	return "", nil

}

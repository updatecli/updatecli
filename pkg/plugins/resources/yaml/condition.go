package yaml

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"gopkg.in/yaml.v3"
)

// Condition checks if a key exists in a yaml file
func (y *Yaml) Condition(source string) (bool, error) {
	return y.condition(source)
}

// ConditionFromSCM checks if a key exists in a yaml file
func (y *Yaml) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	joignedFiles := make(map[string]string)
	for filePath := range y.files {
		joignedFilePath := filePath
		if scm != nil {
			joignedFilePath = joinPathWithWorkingDirectoryPath(joignedFilePath, scm.GetDirectory())
			logrus.Debugf("Relative path detected: changing from %q to absolute path from SCM: %q", filePath, joignedFilePath)
		}
		joignedFiles[joignedFilePath] = y.files[filePath]
	}
	y.files = joignedFiles

	return y.condition(source)
}

func (y *Yaml) condition(source string) (bool, error) {
	var fileContent string
	var filePath string

	if len(y.files) > 1 {
		validationError := fmt.Errorf("Validation error in conditions of type 'yaml': the attributes `spec.files` can't contain more than one element for conditions")
		logrus.Errorf(validationError.Error())
		return false, validationError
	}

	// Start by retrieving the specified file's content
	if err := y.Read(); err != nil {
		return false, err
	}
	out := yaml.Node{}

	// loop over the only file
	for theFilePath := range y.files {
		fileContent = y.files[theFilePath]
		filePath = theFilePath
	}
	err := yaml.Unmarshal([]byte(fileContent), &out)

	if err != nil {
		return false, fmt.Errorf("cannot unmarshal data: %v", err)
	}

	// If a source is provided, then the key 'Value' cannot be specified
	valueToCheck := y.spec.Value

	// When user want to only check the existence of a YAML key
	if y.spec.KeyOnly {
		// Then there must not be any specified Value
		if y.spec.Value != "" {
			validationError := fmt.Errorf("validation error in condition of type 'yaml': both `spec.value` and `spec.keyonly` specified while mutually exclusive. Remove one of these 2 directives")
			logrus.Errorf(validationError.Error())
			return false, validationError
		}

		valueFound, _, _ := replace(&out, strings.Split(y.spec.Key, "."), "", 1)

		return valueFound, nil
	}

	// When user want to check the value of YAML key and when the input source value is not empty
	if source != "" {
		// Then there must not be any specified Value
		if y.spec.Value != "" {
			validationError := fmt.Errorf("validation error in condition of type 'yaml': input source value detected, while `spec.value` specified. Add 'disablesourceinput: true' to your manifest to keep ``spec.value`")
			logrus.Errorf(validationError.Error())
			return false, validationError
		}

		// Use the source input value in this case
		valueToCheck = source
	}

	valueFound, oldVersion, _ := replace(&out, strings.Split(y.spec.Key, "."), valueToCheck, 1)

	if valueFound {
		if oldVersion == valueToCheck {
			logrus.Infof("%s Key %q, in YAML file %q, is correctly set to %q",
				result.SUCCESS,
				y.spec.Key,
				filePath,
				valueToCheck)
			return true, nil
		}

		logrus.Infof("%s Key %q, in YAML file %q, is incorrectly set to %s and should be %q",
			result.FAILURE,
			y.spec.Key,
			filePath,
			oldVersion,
			valueToCheck)
		return false, nil
	}
	return false, fmt.Errorf("%s cannot find key %q in the YAML file %q",
		result.FAILURE,
		y.spec.Key,
		filePath)
}

package yaml

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/scm"
	"gopkg.in/yaml.v3"
)

// Condition checks if a key exists in a yaml file
func (y *Yaml) Condition(source string) (bool, error) {
	return y.condition(source)
}

// ConditionFromSCM checks if a key exists in a yaml file
func (y *Yaml) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {
	if !filepath.IsAbs(y.Spec.File) {
		y.Spec.File = filepath.Join(scm.GetDirectory(), y.Spec.File)
	}
	return y.condition(source)
}

func (y *Yaml) condition(source string) (bool, error) {
	// Start by retrieving the specified file's content
	if err := y.Read(); err != nil {
		return false, err
	}
	out := yaml.Node{}

	err := yaml.Unmarshal([]byte(y.CurrentContent), &out)

	if err != nil {
		return false, fmt.Errorf("cannot unmarshal data: %v", err)
	}

	// If a source is provided, then the key 'Value' cannot be specified
	valueToCheck := y.Spec.Value

	if len(source) > 0 {
		if len(y.Spec.Value) > 0 {
			validationError := fmt.Errorf("Validation error in condition of type 'yaml': the attributes `sourceID` and `spec.value` are mutually exclusive")
			logrus.Errorf(validationError.Error())
			return false, validationError
		}

		valueToCheck = source
	}

	valueFound, oldVersion, _ := replace(&out, strings.Split(y.Spec.Key, "."), valueToCheck, 1)

	if valueFound && oldVersion == valueToCheck {
		logrus.Infof("%s Key %q, in YAML file %q, is correctly set to %q",
			result.SUCCESS,
			y.Spec.Key,
			y.Spec.File,
			valueToCheck)
		return true, nil
	} else if valueFound && oldVersion != valueToCheck {
		logrus.Infof("%s Key %q, in YAML file %q, is incorrectly set to %s and should be %q",
			result.FAILURE,
			y.Spec.Key,
			y.Spec.File,
			oldVersion,
			valueToCheck)
	} else {
		logrus.Infof("%s cannot find key %q in the YAML file %q",
			result.FAILURE,
			y.Spec.Key,
			y.Spec.File)
	}

	return false, nil
}

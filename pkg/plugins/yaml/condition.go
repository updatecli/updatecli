package yaml

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/scm"
	"gopkg.in/yaml.v3"
)

// Condition checks if a key exists in a yaml file
func (y *Yaml) Condition(source string) (bool, error) {
	// Start by retrieving the specified file's content
	if err := y.Read(); err != nil {
		return false, err
	}
	data := y.CurrentContent

	out := yaml.Node{}

	err := yaml.Unmarshal([]byte(data), &out)

	if err != nil {
		return false, fmt.Errorf("cannot unmarshal data: %v", err)
	}

	valueFound, oldVersion, _ := replace(&out, strings.Split(y.Spec.Key, "."), y.Spec.Value, 1)

	if valueFound && oldVersion == y.Spec.Value {
		logrus.Infof("\u2714 Key '%s', from file '%v', is correctly set to %s'",
			y.Spec.Key,
			y.Spec.File,
			y.Spec.Value)
		return true, nil
	} else if valueFound && oldVersion != y.Spec.Value {
		logrus.Infof("\u2717 Key '%s', from file '%v', is incorrectly set to %s and should be %s'",
			y.Spec.Key,
			y.Spec.File,
			oldVersion,
			y.Spec.Value)
	} else {
		logrus.Infof("\u2717 cannot find key '%s' from file '%s'",
			y.Spec.Key,
			y.Spec.File)
	}

	return false, nil
}

// ConditionFromSCM checks if a key exists in a yaml file
func (y *Yaml) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {
	// Start by retrieving the specified file's content
	if err := y.Read(); err != nil {
		return false, err
	}
	data := y.CurrentContent

	out := yaml.Node{}

	err := yaml.Unmarshal([]byte(data), &out)

	if err != nil {
		return false, fmt.Errorf("cannot unmarshal data: %v", err)
	}

	valueFound, oldVersion, _ := replace(&out, strings.Split(y.Spec.Key, "."), y.Spec.Value, 1)

	if valueFound && oldVersion == y.Spec.Value {
		logrus.Infof("\u2714 Key '%s', from file '%v', is correctly set to %s'",
			y.Spec.Key,
			y.Spec.File,
			y.Spec.Value)
		return true, nil
	} else if valueFound && oldVersion != y.Spec.Value {
		logrus.Infof("\u2717 Key '%s', from file '%v', is incorrectly set to %s and should be %s'",
			y.Spec.Key,
			y.Spec.File,
			oldVersion,
			y.Spec.Value)
	} else {
		logrus.Infof("\u2717 cannot find key '%s' from file '%s'",
			y.Spec.Key,
			y.Spec.File)
	}

	return false, nil
}

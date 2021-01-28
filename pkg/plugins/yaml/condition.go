package yaml

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"

	"github.com/olblak/updateCli/pkg/core/scm"
	"github.com/olblak/updateCli/pkg/plugins/file"
	"gopkg.in/yaml.v3"
)

// Condition checks if a key exists in a yaml file
func (y *Yaml) Condition(source string) (bool, error) {

	if len(y.Path) > 0 {
		logrus.Warnf("Key 'Path' is obsolete and now directly defined from file")
	}

	data, err := file.Read(y.File, "")
	if err != nil {
		return false, err
	}

	out := yaml.Node{}

	err = yaml.Unmarshal(data, &out)

	if err != nil {
		return false, fmt.Errorf("cannot unmarshal data: %v", err)
	}

	valueFound, oldVersion, _ := replace(&out, strings.Split(y.Key, "."), y.Value, 1)

	if valueFound && oldVersion == y.Value {
		logrus.Infof("\u2714 Key '%s', from file '%v', is correctly set to %s'\n",
			y.Key,
			y.File,
			y.Value)
		return true, nil
	} else if valueFound && oldVersion != y.Value {
		logrus.Infof("\u2717 Key '%s', from file '%v', is incorrectly set to %s and should be %s'\n",
			y.Key,
			y.File,
			oldVersion,
			y.Value)
	} else {
		logrus.Infof("\u2717 cannot find key '%s' from file '%s'\n",
			y.Key,
			y.File)
	}

	return false, nil
}

// ConditionFromSCM checks if a key exists in a yaml file
func (y *Yaml) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {
	if len(y.Path) > 0 {
		logrus.Warnf("Key 'Path' is obsolete and now directly defined from file")
	}

	data, err := file.Read(y.File, scm.GetDirectory())
	if err != nil {
		return false, err
	}

	out := yaml.Node{}

	err = yaml.Unmarshal(data, &out)

	if err != nil {
		return false, fmt.Errorf("cannot unmarshal data: %v", err)
	}

	valueFound, oldVersion, _ := replace(&out, strings.Split(y.Key, "."), y.Value, 1)

	if valueFound && oldVersion == y.Value {
		logrus.Infof("\u2714 Key '%s', from file '%v', is correctly set to %s'\n",
			y.Key,
			y.File,
			y.Value)
		return true, nil
	} else if valueFound && oldVersion != y.Value {
		logrus.Infof("\u2717 Key '%s', from file '%v', is incorrectly set to %s and should be %s'\n",
			y.Key,
			y.File,
			oldVersion,
			y.Value)
	} else {
		logrus.Infof("\u2717 cannot find key '%s' from file '%s'\n",
			y.Key,
			y.File)
	}

	return false, nil
}

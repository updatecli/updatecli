package yaml

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/olblak/updateCli/pkg/core/scm"
	"gopkg.in/yaml.v3"
)

// Condition checks if a key exists in a yaml file
func (y *Yaml) Condition(source string) (bool, error) {

	// By default workingDir is set to local directory
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	workingDir := filepath.Dir(pwd)

	if dir, base, err := isFileExist(y.File); err == nil && y.Path == "" {
		// if no scm configuration has been provided and neither file path then we try to guess the file directory.
		// if file name contains a path then we use it otherwise we fallback to the current path
		y.Path = dir
		y.File = base
	} else if _, _, err := isFileExist(y.File); err != nil && y.Path == "" {

		y.Path = workingDir

	} else if y.Path != "" && !isDirectory(y.Path) {

		fmt.Printf("Directory '%s' is not valid so fallback to '%s'", y.Path, workingDir)
		y.Path = workingDir

	} else {
		return false, fmt.Errorf("Something weird happened while trying to set working directory")
	}

	exist := false

	data, err := y.ReadFile()
	if err != nil {
		return exist, err
	}

	var out yaml.Node

	err = yaml.Unmarshal(data, &out)

	if err != nil {
		return exist, fmt.Errorf("cannot unmarshal data: %v", err)
	}

	valueFound, oldVersion, _ := replace(&out, strings.Split(y.Key, "."), y.Value, 1)

	if valueFound && oldVersion == y.Value {
		exist = true
		fmt.Printf("\u2714 Key '%s', from file '%v', is correctly set to %s'\n",
			y.Key,
			filepath.Join(y.Path, y.File),
			y.Value)

	} else if valueFound && oldVersion != y.Value {

		fmt.Printf("\u2717 Key '%s', from file '%v', is incorrectly set to %s and should be %s'\n",
			y.Key,
			filepath.Join(y.Path, y.File),
			oldVersion,
			y.Value)

	} else {
		fmt.Printf("\u2717 cannot find key '%s' from file '%s'\n",
			y.Key,
			filepath.Join(y.Path, y.File))

		return exist, nil
	}

	return exist, nil
}

// ConditionFromSCM checks if a key exists in a yaml file
func (y *Yaml) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {

	y.Path = scm.GetDirectory()

	exist := false

	data, err := y.ReadFile()
	if err != nil {
		return exist, err
	}

	var out yaml.Node

	err = yaml.Unmarshal(data, &out)

	if err != nil {
		return exist, fmt.Errorf("cannot unmarshal data: %v", err)
	}

	valueFound, oldVersion, _ := replace(&out, strings.Split(y.Key, "."), y.Value, 1)

	if valueFound && oldVersion == y.Value {
		exist = true
		fmt.Printf("\u2714 Key '%s', from file '%v', is correctly set to %s'\n",
			y.Key,
			filepath.Join(y.Path, y.File),
			y.Value)

	} else if valueFound && oldVersion != y.Value {
		fmt.Printf("\u2717 Key '%s', from file '%v', is incorrectly set to %s and should be %s'\n",
			y.Key,
			filepath.Join(y.Path, y.File),
			oldVersion,
			y.Value)

	} else {
		fmt.Printf("\u2717 cannot find key '%s' from file '%s'\n",
			y.Key,
			filepath.Join(y.Path, y.File))

		return exist, nil
	}

	exist = true

	return exist, nil
}

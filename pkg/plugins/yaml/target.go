package yaml

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/olblak/updateCli/pkg/core/scm"
	"gopkg.in/yaml.v3"
)

// Target updates a scm repository based on the modified yaml file.
func (y *Yaml) Target(source string, dryRun bool) (changed bool, err error) {

	y.Value = source

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

	changed = false

	data, err := y.ReadFile()

	if err != nil {
		return changed, err
	}

	var out yaml.Node

	err = yaml.Unmarshal(data, &out)

	if err != nil {
		return changed, fmt.Errorf("cannot unmarshal data: %v", err)
	}

	valueFound, oldVersion, _ := replace(&out, strings.Split(y.Key, "."), y.Value, 1)

	if valueFound {
		if oldVersion == y.Value {
			fmt.Printf("\u2714 Key '%s', from file '%v', already set to %s, nothing else need to be done\n",
				y.Key,
				filepath.Join(y.Path, y.File),
				y.Value)
			return changed, nil
		}

		fmt.Printf("\u2714 Key '%s', from file '%v', was updated from '%s' to '%s'\n",
			y.Key,
			filepath.Join(y.Path, y.File),
			oldVersion,
			y.Value)

	} else {
		fmt.Printf("\u2717 cannot find key '%s' from file '%s'\n", y.Key, y.Path)
		return changed, nil
	}

	if !dryRun {

		newFile, err := os.Create(filepath.Join(y.Path, y.File))
		defer newFile.Close()

		encoder := yaml.NewEncoder(newFile)
		defer encoder.Close()
		encoder.SetIndent(yamlIdent)
		err = encoder.Encode(&out)

		if err != nil {
			return changed, fmt.Errorf("something went wrong while encoding %v", err)
		}
	}

	changed = true

	return changed, nil
}

// TargetFromSCM updates a scm repository based on the modified yaml file.
func (y *Yaml) TargetFromSCM(source string, scm scm.Scm, dryRun bool) (changed bool, files []string, message string, err error) {

	y.Path = scm.GetDirectory()
	y.Value = source

	changed = false

	data, err := y.ReadFile()

	if err != nil {
		return changed, files, message, err
	}

	var out yaml.Node

	err = yaml.Unmarshal(data, &out)

	if err != nil {
		return changed, files, message, fmt.Errorf("cannot unmarshal data: %v", err)
	}

	valueFound, oldVersion, _ := replace(&out, strings.Split(y.Key, "."), y.Value, 1)

	if valueFound {
		if oldVersion == y.Value {
			fmt.Printf("\u2714 Key '%s', from file '%v', already set to %s, nothing else need to be done\n",
				y.Key,
				filepath.Join(y.Path, y.File),
				y.Value)
			return changed, files, message, nil
		}

		fmt.Printf("\u2714 Key '%s', from file '%v', was updated from '%s' to '%s'\n",
			y.Key,
			filepath.Join(y.Path, y.File),
			oldVersion,
			y.Value)

	} else {
		fmt.Printf("\u2717 cannot find key '%s' from file '%s'\n", y.Key, y.Path)
		return changed, files, message, nil
	}

	if !dryRun {

		newFile, err := os.Create(filepath.Join(y.Path, y.File))
		defer newFile.Close()

		encoder := yaml.NewEncoder(newFile)
		defer encoder.Close()
		encoder.SetIndent(yamlIdent)
		err = encoder.Encode(&out)

		if err != nil {
			return changed, files, message, fmt.Errorf("something went wrong while encoding %v", err)
		}
	}

	files = append(files, y.File)
	message = fmt.Sprintf("[updatecli] Key '%s', from file '%v', was updated from %s to '%s'\n",
		y.Key,
		y.File,
		oldVersion,
		y.Value)

	changed = true

	return changed, files, message, nil
}

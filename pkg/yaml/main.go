package yaml

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	yamlIdent int = 2
)

// Yaml stores configuration about the file and the key value which needs to be updated.
type Yaml struct {
	File string
	Key  string
}

// GetFile returns the yaml file which need to be updated.
func (y *Yaml) GetFile() string {
	return y.File
}

// replace parses a yaml object looking for a specific key which needs to be updated and replace it if needed.
func replace(entry *yaml.Node, keys []string, version string, columnRef int) (found bool, oldVersion string, column int) {

	valueFound := false
	column = columnRef
	nextLevel := false
	for _, content := range entry.Content {
		if content.Column < column {
			break
		}
		if nextLevel {
			column = content.Column
			nextLevel = false
		}

		if content.Value == keys[0] && (content.Column == columnRef) {
			column = content.Column
			nextLevel = true

			if len(keys) > 1 {
				keys = keys[1:]
			} else if len(keys) == 1 {
				valueFound = true
				continue
			}
		}

		if content.Kind == yaml.ScalarNode && valueFound == true {
			column = content.Column

			oldVersion = content.Value
			content.SetString(version)

			break
		} else if content.Kind == yaml.MappingNode {
			valueFound, oldVersion, column = replace(content, keys, version, column)
		}
	}
	return valueFound, oldVersion, column
}

// Target updates a scm repository based on the modified yaml file.
func (y *Yaml) Target(source string, workDir string) (message string, err error) {

	fmt.Printf("\nUpdating key  '%s' from target file: %s:\n\n", y.Key, y.File)

	path := filepath.Join(workDir, y.File)

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	var out yaml.Node

	err = yaml.Unmarshal(data, &out)

	if err != nil {
		return "", fmt.Errorf("cannot unmarshal data: %v", err)
	}

	valueFound, oldVersion, _ := replace(&out, strings.Split(y.Key, "."), source, 1)

	if valueFound {
		if oldVersion == source {
			fmt.Printf("\u2714 Value '%s' already up to date\n", source)
			return "", nil
		}

		fmt.Printf("\u2717 Version mismatched between %s (old) and %s (new)\n", oldVersion, source)

	} else {
		fmt.Printf("\u2717 cannot find key '%v' in file %v\n", y.Key, path)
		return "", nil
	}

	message = fmt.Sprintf("[updatecli] %s - Updating key '%v' to %s",
		y.File,
		y.Key,
		source)

	newFile, err := os.Create(path)
	defer newFile.Close()

	encoder := yaml.NewEncoder(newFile)
	defer encoder.Close()
	encoder.SetIndent(yamlIdent)
	err = encoder.Encode(&out)

	if err != nil {
		return "", fmt.Errorf("something went wrong while encoding %v", err)
	}

	return message, nil
}

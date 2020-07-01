package yaml

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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

// isPositionKey checks if key use the array position format
func isPositionKey(key string) bool {
	matched, err := regexp.MatchString("^[[:alnum:]]*[[[:digit:]]*]$", key)

	if err != nil {
		fmt.Println(err)
	}
	return matched
}

func getPositionKeyValue(k string) (key string, position int, err error) {
	if isPositionKey(k) {
		re := regexp.MustCompile(`^([[:alnum:]]*)\[[[:digit:]]*\]$`)
		keys := re.FindStringSubmatch(k)
		key = keys[1]

		re = regexp.MustCompile(`^[[:alnum:]]*\[([[:digit:]]*)\]$`)

		positions := re.FindStringSubmatch(k)

		fmt.Println(positions)
		position, err = strconv.Atoi(positions[1])

		if err != nil {
			fmt.Println(err)
			return "", -1, err
		}

	} else {
		if strings.ContainsAny(k, "_?:,[]{}#&*!|>`\"%") {
			key = ""
			position = -1
			err = fmt.Errorf("Error: key '%s' cannot contains yaml special characters", k)

		} else {
			key = k
			position = -1
			err = nil
		}
	}

	return key, position, err

}

// replace parses a yaml object looking for a specific key which needs to be updated and replace it if needed.
func replace(entry *yaml.Node, keys []string, version string, columnRef int) (found bool, oldVersion string, column int) {
	/*
		In yaml.MappingNodes, child nodes represent yaml keys when index is even, and yaml values when index is odd
	*/

	valueFound := false
	column = columnRef

	key, position, err := getPositionKeyValue(keys[0])
	if err != nil {
		fmt.Println(err)
	}

	// If document start with a sequence and we are looking for an array position
	if entry.Kind == yaml.DocumentNode &&
		entry.Content[0].Kind == yaml.SequenceNode &&
		isPositionKey(keys[0]) {

		if entry.Content[0].Content[0].Kind == yaml.MappingNode {
			if len(keys) > 1 {
				positionIndex := 0

				if (position*2)-1 >= 0 {
					positionIndex = (position * 2) - 1
				}

				if len(entry.Content[0].Content) < positionIndex {
					return false, "", 0
				}

				keys = keys[1:]

				column = entry.Content[0].Content[positionIndex].Column

				valueFound, oldVersion, column = replace(entry.Content[0].Content[positionIndex],
					keys,
					version,
					entry.Content[0].Content[positionIndex].Column)
			} else {

				if len(entry.Content[0].Content) < position {
					return false, "", 0
				}

				valueFound = true
				oldVersion = entry.Content[0].Content[position].Value
				column = entry.Content[0].Content[position].Column
			}

		} else if entry.Content[0].Content[0].Kind == yaml.ScalarNode && len(keys) == 1 {

			if len(entry.Content[0].Content) < position {
				return false, "", 0
			}

			oldVersion = entry.Content[0].Content[position].Value
			valueFound = true
			column = entry.Content[0].Content[position].Column

		} else {
			return false, "", 0
		}

		return valueFound, oldVersion, column

	}

	for index, content := range entry.Content {
		// In yaml.MappingNodes, child nodes represent yaml keys when index is even and yaml values when index is odd
		if index%2 == 0 && content.Value == key && (content.Column == columnRef) {

			if len(keys) > 1 {
				keys = keys[1:]
				column = entry.Content[index+1].Column

				if entry.Content[index+1].Kind == yaml.SequenceNode && entry.Content[index+1].Content[0].Kind == yaml.MappingNode {
					positionIndex := 0

					if (position*2)-1 >= 0 {
						positionIndex = (position * 2) - 1
					}

					if positionIndex > len(entry.Content[index+1].Content) {
						return false, "", 0
					}

					valueFound, oldVersion, column = replace(entry.Content[index+1].Content[positionIndex],
						keys,
						version,
						entry.Content[index+1].Content[positionIndex].Column)
				}

				key, position, err = getPositionKeyValue(keys[0])
				if err != nil {
					fmt.Println(err)
				}
			} else {
				if entry.Content[index+1].Kind == yaml.ScalarNode {
					column = entry.Content[index+1].Column

					oldVersion = entry.Content[index+1].Value
					entry.Content[index+1].SetString(version)
					valueFound = true
					break

				} else if entry.Content[index+1].Kind == yaml.SequenceNode {

					if len(entry.Content[index+1].Content) < position {
						return false, "", 0
					}

					oldVersion = entry.Content[index+1].Content[position].Value
					valueFound = true
					column = entry.Content[index+1].Content[position].Column

					break
				}
			}
			continue
		}
		if content.Kind == yaml.MappingNode {
			valueFound, oldVersion, column = replace(content, keys, version, column)
		}
		if content.Column < column {
			break
		}
	}
	return valueFound, oldVersion, column
}

// Target updates a scm repository based on the modified yaml file.
func (y *Yaml) Target(source string, name string, workDir string) (changed bool, message string, err error) {

	changed = false

	path := filepath.Join(workDir, y.File)

	file, err := os.Open(path)
	if err != nil {
		return changed, "", err
	}

	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return changed, "", err
	}

	var out yaml.Node

	err = yaml.Unmarshal(data, &out)

	if err != nil {
		return changed, "", fmt.Errorf("cannot unmarshal data: %v", err)
	}

	valueFound, oldVersion, _ := replace(&out, strings.Split(y.Key, "."), source, 1)

	if valueFound {
		if oldVersion == source {
			fmt.Printf("\u2714 Key '%s', from file '%v', already set to %s, nothing else need to be done\n",
				y.Key,
				y.File,
				source)
			return changed, "", nil
		}

		fmt.Printf("\u2714 Key '%s', from file '%v', was updated from '%s' to '%s'\n",
			y.Key,
			y.File,
			oldVersion,
			source)

	} else {
		fmt.Printf("\u2717 cannot find key '%s' from file '%s'\n", y.Key, path)
		return changed, "", nil
	}

	message = fmt.Sprintf("[updatecli] Update %s version to %v\n\nKey '%s', from file '%v', was updated to '%s'\n",
		name,
		source,
		y.Key,
		y.File,
		source)

	newFile, err := os.Create(path)
	defer newFile.Close()

	encoder := yaml.NewEncoder(newFile)
	defer encoder.Close()
	encoder.SetIndent(yamlIdent)
	err = encoder.Encode(&out)

	if err != nil {
		return changed, "", fmt.Errorf("something went wrong while encoding %v", err)
	}

	changed = true

	return changed, message, nil
}

package yaml

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/text"

	"gopkg.in/yaml.v3"
)

var (
	yamlIdent int = 2
)

// Spec defines a specification for a "yaml" resource
// parsed from an updatecli manifest file
type Spec struct {
	File string
	Key  string
	// Deprecated: use File instead
	Path  string
	Value string
}

// Yaml defines a resource of kind "yaml"
type Yaml struct {
	spec             Spec
	contentRetriever text.TextRetriever
	currentContent   string
}

// New returns a reference to a newly initialized Yaml object from a YamlSpec
// or an error if the provided YamlSpec triggers a validation error.
func New(newSpec Spec) (*Yaml, error) {

	newResource := &Yaml{
		spec:             newSpec,
		contentRetriever: &text.Text{},
	}
	// TODO: generalize the Validate + Normalize as an interface to all resources
	err := newResource.Validate()
	if err != nil {
		return nil, err
	}

	newResource.spec.File = strings.TrimPrefix(newResource.spec.File, "file://")

	return newResource, nil

}

// Validate validates the object and returns an error (with all the failed validation messages) if it is not valid
func (y *Yaml) Validate() error {
	var validationErrors []string

	// Check for all validation
	if len(y.spec.Path) > 0 {
		validationErrors = append(validationErrors, "Invalid spec for yaml resource: Key 'path' is deprecated, use 'file' instead.")
	}
	if y.spec.File == "" {
		validationErrors = append(validationErrors, "Invalid spec for yaml resource: 'file' is empty.")
	} else {
		if !y.contentRetriever.FileExists(y.spec.File) {
			validationErrors = append(validationErrors, fmt.Sprintf("Invalid spec for yaml resource: the file %q does not exist.", y.spec.File))
		}
	}
	if y.spec.Key == "" {
		validationErrors = append(validationErrors, "Invalid spec for yaml resource: 'key' is empty.")
	}
	// Return all the validation errors if found any
	if len(validationErrors) > 0 {
		return fmt.Errorf("Validation error: the provided manifest configuration had the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	return nil
}

// Read defines CurrentContent to the content of the file which path is specified in Spec.File
func (y *Yaml) Read() error {
	// Otherwise return the textual content
	textContent, err := y.contentRetriever.ReadAll(y.spec.File)
	if err != nil {
		return err
	}
	y.currentContent = textContent

	return nil
}

// isPositionKey checks if key use the array position like agents[0].
// where [0] references the position 0 in the array 'agents'
func isPositionKey(key string) bool {
	matched, err := regexp.MatchString("(.*)[[[:digit:]]+]$", key)

	if err != nil {
		logrus.Errorf("err - %s", err)
	}
	return matched
}

func getPositionKeyValue(k string) (key string, position int, err error) {
	if isPositionKey(k) {
		re := regexp.MustCompile(`^(.*)\[[[:digit:]]*\]$`)
		keys := re.FindStringSubmatch(k)
		key = keys[1]

		re = regexp.MustCompile(`^.*\[([[:digit:]]*)\]$`)

		positions := re.FindStringSubmatch(k)

		position, err = strconv.Atoi(positions[1])

		if err != nil {
			logrus.Errorf("err - %s", err)
			return "", -1, err
		}

	} else {
		key = k
		position = -1
		err = nil
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
		logrus.Errorf("err - %s", err)
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
					/*
						If looking for a key at a specific position in a sequence like keys[2].key
						for following example should parse - key : c
						keys:
							- key: a
							- key: b
							- key: c
					*/

					positionIndex := position

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
					logrus.Errorf("err - %s", err)
				}
			} else {
				if entry.Content[index+1].Kind == yaml.ScalarNode && !isPositionKey(keys[0]) {
					/*
						It means we reached the key/value
						that need to be updated like :
						array:
							- name: key0
							- name: key1
					*/

					column = entry.Content[index+1].Column

					oldVersion = entry.Content[index+1].Value
					entry.Content[index+1].SetString(version)
					valueFound = true
					break

				} else if entry.Content[index+1].Kind == yaml.SequenceNode && isPositionKey(keys[0]) {
					/*
						It means we reached the list of values
						that need to be updated like :
						array:
							- key0
							- key1
					*/

					if len(entry.Content[index+1].Content) < position {
						return false, "", 0
					}

					oldVersion = entry.Content[index+1].Content[position].Value
					entry.Content[index+1].Content[position].SetString(version)
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

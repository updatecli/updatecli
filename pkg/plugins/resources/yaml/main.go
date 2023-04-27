package yaml

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"

	"gopkg.in/yaml.v3"
)

/*
"yaml"  defines the specification for manipulating "yaml" files.
It can be used as a "source", a "condition", or a "target".
*/
type Spec struct {
	/*
		"file" defines the yaml file path to interact with.

		compatible:
			* source
			* condition
			* target

		remark:
			* "file" and "files" are mutually exclusive
			* when used as a source or condition, the file path also accept the following protocols
			* protocols "https://", "http://", and "file://" are supported in path for source and condition
	*/
	File string `yaml:",omitempty"`
	/*
		"files" defines the list of yaml files path to interact with.

		compatible:
			* condition
			* target

		remark:
			* file and files are mutually exclusive
			* protocols "https://", "http://", and "file://" are supported in file path for source and condition
	*/
	Files []string `yaml:",omitempty"`
	/*
		"key" defines the yaml key used to identify the yaml information to manipulate.

		compatible:
			* source
			* condition
			* target

		remark:
			* key is a simpler version of yamlpath accepts keys.

		example:
			* key: name
			* key: agent.name
			* key: agents[0].name

	*/
	Key string `yaml:",omitempty"`
	/*
		"value" is the value associated with a yaml key.

		compatible:
			* source
			* condition
			* target

		default:
			When used from a condition or a target, the default value is set to linked source output.
	*/
	Value string `yaml:",omitempty"`
	/*
		"keyonly" allows to only check if a key exist and do not return an error otherwise

		compatible:
			* condition

		default:
			false
	*/
	KeyOnly bool `yaml:",omitempty"`
	/*
		"indent" allows to override the default yaml indentation for file updated by a target.

		compatible:
			* target

		default:
			default indentation is set to 2.
	*/
	Indent int `yaml:",omitempty"`
}

// Yaml defines a resource of kind "yaml"
type Yaml struct {
	spec             Spec
	contentRetriever text.TextRetriever
	indent           int
	files            map[string]file // map of file paths to file contents
}

type file struct {
	originalFilePath string
	filePath         string
	content          string
}

// New returns a reference to a newly initialized Yaml object from a Spec
// or an error if the provided YamlSpec triggers a validation error.
func New(spec interface{}) (*Yaml, error) {
	newSpec := Spec{}

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return nil, err
	}

	newResource := &Yaml{
		spec:             newSpec,
		contentRetriever: &text.Text{},
	}

	err = newResource.spec.Validate()
	if err != nil {
		return nil, err
	}

	newResource.indent = newResource.spec.Indent
	if newResource.indent <= 0 {
		newResource.indent = 2
	}

	newResource.files = make(map[string]file)
	// File as unique element of newResource.files
	if len(newResource.spec.File) > 0 {
		filePath := strings.TrimPrefix(newResource.spec.File, "file://")
		newResource.files[filePath] = file{
			originalFilePath: filePath,
			filePath:         filePath,
		}
	}
	// Files
	for _, filePath := range newResource.spec.Files {
		filePath := strings.TrimPrefix(filePath, "file://")
		newResource.files[strings.TrimPrefix(filePath, "file://")] = file{
			originalFilePath: filePath,
			filePath:         filePath,
		}
	}

	return newResource, nil
}

func hasDuplicates(values []string) bool {
	uniqueValues := make(map[string]string)
	for _, v := range values {
		uniqueValues[v] = ""
	}

	return len(values) != len(uniqueValues)
}

// Validate validates the object and returns an error (with all the failed validation messages) if it is not valid
func (s *Spec) Validate() error {
	var validationErrors []string

	// Check for all validation
	if len(s.Files) == 0 && s.File == "" {
		validationErrors = append(validationErrors, "Invalid spec for yaml resource: both 'file' and 'files' are empty.")
	}
	if s.Key == "" {
		validationErrors = append(validationErrors, "Invalid spec for yaml resource: 'key' is empty.")
	}
	if len(s.Files) > 0 && s.File != "" {
		validationErrors = append(validationErrors, "Validation error in target of type 'yaml': the attributes `spec.file` and `spec.files` are mutually exclusive")
	}
	if len(s.Files) > 1 && hasDuplicates(s.Files) {
		validationErrors = append(validationErrors, "Validation error in target of type 'yaml': the attributes `spec.files` contains duplicated values")
	}

	if s.Indent < 0 {
		validationErrors = append(validationErrors, "Invalid spec for yaml resource: indentation must be a >= 0")
	}
	// Return all the validation errors if found any
	if len(validationErrors) > 0 {
		return fmt.Errorf("validation error: the provided manifest configuration had the following validation errors:\n%s", strings.Join(validationErrors, "\n\n"))
	}

	return nil
}

// Read puts the content of the file(s) as value of the y.files map if the file(s) exist(s) or log the non existence of the file
func (y *Yaml) Read() error {
	var err error

	// Retrieve files content
	for filePath := range y.files {
		f := y.files[filePath]
		if y.contentRetriever.FileExists(f.filePath) {
			f.content, err = y.contentRetriever.ReadAll(f.filePath)
			if err != nil {
				return err
			}
			y.files[filePath] = f

		} else {
			return fmt.Errorf("%s The specified file %q does not exist", result.FAILURE, f.filePath)
		}
	}
	return nil
}

func (y *Yaml) UpdateAbsoluteFilePath(workDir string) {
	for filePath := range y.files {
		if workDir != "" {
			f := y.files[filePath]
			f.filePath = joinPathWithWorkingDirectoryPath(f.originalFilePath, workDir)

			logrus.Debugf("Relative path detected: changing from %q to absolute path from SCM: %q", f.originalFilePath, f.filePath)
			y.files[filePath] = f
		}
	}
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

					if positionIndex >= len(entry.Content[index+1].Content) {
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

// Changelog returns the changelog for this resource, or an empty string if not supported
func (y *Yaml) Changelog() string {
	return ""
}

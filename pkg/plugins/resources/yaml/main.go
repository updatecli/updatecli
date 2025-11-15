package yaml

import (
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
)

/*
"yaml"  defines the specification for manipulating "yaml" files.
It can be used as a "source", a "condition", or a "target".
*/
type Spec struct {
	// DocumentIndex defines the index of the document to interact with in a multi-document yaml file.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// default:
	//   empty
	//
	//  remark:
	//   * when not set, all documents are considered
	//   * Only compatible with the "go-yaml" engine
	//
	// example:
	//   * documentindex: 0
	//   * documentindex: 1
	//
	DocumentIndex *int `yaml:",omitempty"`
	//"engine" defines the engine to use to manipulate the yaml file.
	//
	//There is no one good Golang library to manipulate yaml files.
	//And each one of them have has its pros and cons so we decided to allow this customization based on user's needs.
	//
	//remark:
	//  * Accepted value is one of "yamlpath", "go-yaml","default" or nothing
	//  * go-yaml, "default" and "" are equivalent
	Engine string `yaml:",omitempty"`
	//"file" defines the yaml file path to interact with.
	//
	//compatible:
	//  * source
	//  * condition
	//  * target
	//
	//remark:
	//  * "file" and "files" are mutually exclusive
	//  * scheme "https://", "http://", and "file://" are supported in path for source and condition
	//
	File string `yaml:",omitempty"`
	//"files" defines the list of yaml files path to interact with.
	//
	//compatible:
	//  * condition
	//  * target
	//
	//remark:
	//  * file and files are mutually exclusive
	//  * protocols "https://", "http://", and "file://" are supported in file path for source and condition
	//
	Files []string `yaml:",omitempty"`
	//"key" defines the yaml keypath.
	//
	//compatible:
	//  * source
	//  * condition
	//  * target
	//
	//remark:
	//  * key is a simpler version of yamlpath accepts keys.
	//
	//example using default engine:
	//  * key: $.name
	//  * key: $.agent.name
	//  * key: $.agents[0].name
	//  * key: $.agents[*].name
	//  * key: $.'agents.name'
	//  * key: $.repos[?(@.repository == 'website')].owner" (require engine set to yamlpath)
	//
	//remark:
	//  field path with key/value is not supported at the moment.
	//  some help would be useful on https://github.com/goccy/go-yaml/issues/290
	//
	Key string `yaml:",omitempty"`
	//"keys" defines multiple yaml keypaths to update with the same value.
	//
	//compatible:
	//  * target
	//
	//remark:
	//  * keys is mutually exclusive with key.
	//  * keys accepts the same syntax as key for each element.
	//  * all keys will be updated with the same value.
	//  * only available for target operations, not for source or condition.
	//
	//example using default engine:
	//  * keys:
	//    - $.image.tag
	//    - $.sidecar.tag
	//  * keys:
	//    - $.agents[0].version
	//    - $.agents[1].version
	//
	Keys []string `yaml:",omitempty"`
	//value is the value associated with a yaml key.
	//
	//compatible:
	//  * source
	//  * condition
	//  * target
	//
	//default:
	//	When used from a condition or a target, the default value is set to the associated source output.
	//
	Value string `yaml:",omitempty"`
	//keyonly allows to check only if a key exist and do not return an error otherwise
	//
	//compatible:
	//	* condition
	//
	//default:
	//	false
	//
	KeyOnly bool `yaml:",omitempty"`
	//searchpattern defines if the MatchPattern should be applied on the file(s) path
	//
	//If set to true, it modifies the behavior of the `file` and `files` attributes to search for files matching the pattern instead of searching for files with the exact name.
	//When looking for file path pattern, it requires pattern to match all of name, not just a substring.
	//
	//The pattern syntax is:
	//
	//```
	//    pattern:
	//        { term }
	//    term:
	//        '*'         matches any sequence of non-Separator characters
	//        '?'         matches any single non-Separator character
	//        '[' [ '^' ] { character-range } ']'
	//                    character class (must be non-empty)
	//        c           matches character c (c != '*', '?', '\\', '[')
	//        '\\' c      matches character c
	//
	//    character-range:
	//        c           matches character c (c != '\\', '-', ']')
	//        '\\' c      matches character c
	//        lo '-' hi   matches character c for lo <= c <= hi
	//```
	//
	SearchPattern bool `yaml:",omitempty"`
	//comment defines a comment to add after the value.
	//
	//default: empty
	//
	//compatible:
	//  * target
	//
	//remarks:
	//  * Please note that the comment is added if the value is modified by Updatecli
	//
	Comment string `yaml:",omitempty"`
}

// Yaml defines a resource of kind "yaml"
type Yaml struct {
	spec             Spec
	contentRetriever text.TextRetriever
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

	if newResource.spec.Key != "" {
		newResource.spec.Key = sanitizeYamlPathKey(newResource.spec.Key)
	}

	// Sanitize all keys in the Keys slice
	for i, key := range newResource.spec.Keys {
		newResource.spec.Keys[i] = sanitizeYamlPathKey(key)
	}

	err = newResource.spec.Validate()
	if err != nil {
		return nil, err
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

// getKeys returns all keys to be processed, handling both single key and multiple keys
func (s *Spec) getKeys() []string {
	if s.Key != "" {
		return []string{s.Key}
	}
	return s.Keys
}

// Validate validates the object and returns an error (with all the failed validation messages) if it is not valid
func (s *Spec) Validate() error {
	var validationErrors []string

	// Check for all validation
	if len(s.Files) == 0 && s.File == "" {
		validationErrors = append(validationErrors, "Invalid spec for yaml resource: both 'file' and 'files' are empty.")
	}
	if s.Key == "" && len(s.Keys) == 0 {
		validationErrors = append(validationErrors, "Invalid spec for yaml resource: both 'key' and 'keys' are empty.")
	}
	if s.Key != "" && len(s.Keys) > 0 {
		validationErrors = append(validationErrors, "Invalid spec for yaml resource: 'key' and 'keys' are mutually exclusive.")
	}
	if len(s.Files) > 0 && s.File != "" {
		validationErrors = append(validationErrors, "Validation error in target of type 'yaml': the attributes `spec.file` and `spec.files` are mutually exclusive")
	}
	if len(s.Files) > 1 && hasDuplicates(s.Files) {
		validationErrors = append(validationErrors, "Validation error in target of type 'yaml': the attributes `spec.files` contains duplicated values")
	}
	if len(s.Keys) > 1 && hasDuplicates(s.Keys) {
		validationErrors = append(validationErrors, "Validation error in target of type 'yaml': the attribute `spec.keys` contains duplicated values")
	}

	if s.DocumentIndex != nil {
		switch s.Engine {
		case EngineDefault, EngineGoYaml, "":
			//
		default:
			validationErrors = append(validationErrors, "Validation error in target of type 'yaml': the attribute `spec.documentindex` is only compatible with the `go-yaml` engine")
		}
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

// Changelog returns the changelog for this resource, or an empty string if not supported
func (y *Yaml) Changelog(from, to string) *result.Changelogs {
	return nil
}

// initFiles initializes the f.files map
func (y *Yaml) initFiles(workDir string) error {
	y.files = make(map[string]file)

	// File as unique element of newResource.files
	if len(y.spec.File) > 0 {
		var foundFiles []string
		var err error
		switch y.spec.SearchPattern {
		case true:
			foundFiles, err = utils.FindFilesMatchingPathPattern(workDir, y.spec.File)
			if err != nil {
				return fmt.Errorf("unable to find file matching %q: %s", y.spec.File, err)
			}
		case false:
			foundFiles = append(foundFiles, y.spec.File)
		}

		for _, filePath := range foundFiles {
			newFile := file{
				filePath:         strings.TrimPrefix(filePath, "file://"),
				originalFilePath: strings.TrimPrefix(filePath, "file://"),
			}
			y.files[filePath] = newFile
		}
	}

	for _, specFile := range y.spec.Files {
		var foundFiles []string
		var err error

		switch y.spec.SearchPattern {
		case true:
			foundFiles, err = utils.FindFilesMatchingPathPattern(workDir, specFile)
			if err != nil {
				return fmt.Errorf("unable to find files matching %q: %s", y.spec.File, err)
			}

		case false:
			foundFiles = append(foundFiles, y.spec.Files...)
		}

		for _, filePath := range foundFiles {
			newFile := file{
				filePath:         strings.TrimPrefix(filePath, "file://"),
				originalFilePath: strings.TrimPrefix(filePath, "file://"),
			}
			y.files[filePath] = newFile
		}
	}

	for filePath := range y.files {
		if workDir != "" {
			file := y.files[filePath]
			file.filePath = joinPathWithWorkingDirectoryPath(file.originalFilePath, workDir)

			logrus.Debugf("Relative path detected: changing from %q to absolute path from SCM: %q", file.originalFilePath, file.filePath)
			y.files[filePath] = file
		}
	}

	return nil
}

func (y *Yaml) UpdateAbsoluteFilePath(workDir string) {
	for filePath := range y.files {
		if workDir != "" {
			file := y.files[filePath]
			file.filePath = joinPathWithWorkingDirectoryPath(file.originalFilePath, workDir)

			logrus.Debugf("Relative path detected: changing from %q to absolute path from SCM: %q", file.originalFilePath, file.filePath)
			y.files[filePath] = file
		}
	}
}

// ReportConfig returns a new configuration object with only the necessary fields
// to identify the resource without any sensitive information or context specific data.
func (y *Yaml) ReportConfig() interface{} {
	return Spec{
		File:          y.spec.File,
		Files:         y.spec.Files,
		Key:           y.spec.Key,
		Keys:          y.spec.Keys,
		Value:         y.spec.Value,
		Engine:        y.spec.Engine,
		KeyOnly:       y.spec.KeyOnly,
		SearchPattern: y.spec.SearchPattern,
	}
}

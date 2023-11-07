package yaml

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
)

/*
"yaml"  defines the specification for manipulating "yaml" files.
It can be used as a "source", a "condition", or a "target".
*/
type Spec struct {
	/*
		"engine" defines the engine to use to manipulate the yaml file.

		There is no one good Golang library to manipulate yaml files.
		And each one of them have has its pros and cons so we decided to allow this customization based on user's needs.

		remark:
			* Accepted value is one of "yamlpath", "go-yaml","default" or nothing
			* go-yaml, "default" and "" are equivalent
	*/
	Engine string `yaml:",omitempty"`
	/*
		"file" defines the yaml file path to interact with.

		compatible:
			* source
			* condition
			* target

		remark:
			* "file" and "files" are mutually exclusive
			* scheme "https://", "http://", and "file://" are supported in path for source and condition
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
		"key" defines the yaml keypath.

		compatible:
			* source
			* condition
			* target

		remark:
			* key is a simpler version of yamlpath accepts keys.

		example using default engine:
			* key: $.name
			* key: $.agent.name
			* key: $.agents[0].name
			* key: $.agents[*].name
			* key: $.'agents.name'
			* key: $.repos[?(@.repository == 'website')].owner" (require engine set to yamlpath)

		remark:
			field path with key/value is not supported at the moment.
			some help would be useful on https://github.com/goccy/go-yaml/issues/290

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

	newResource.spec.Key = sanitizeYamlPathKey(newResource.spec.Key)

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

// Changelog returns the changelog for this resource, or an empty string if not supported
func (y *Yaml) Changelog() string {
	return ""
}

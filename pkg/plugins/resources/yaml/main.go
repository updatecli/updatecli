package yaml

import (
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

/*
"yaml" defines the specification for manipulating yaml files.
It can be used as a "source", a "condition", or a "target".
*/
type Spec struct {
	// "documentindex" defines the index of the document to use in a multi document yaml file.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// default:
	//   empty
	//
	// remark:
	//   * when unset in a source, the value is retrieved from the first document matching the query.
	//   * when unset in a condition or a target, every document is evaluated by the query.
	//
	// example:
	//   * documentindex: 0
	//   * documentindex: 1
	//
	DocumentIndex *int `yaml:",omitempty"`
	// "engine" defines the library used to manipulate the yaml file.
	//
	// No single Go library handles yaml well in every case, and each has its own strengths,
	// so Updatecli lets you pick the one that suits your file.
	//
	// default:
	//   go-yaml
	//
	// remark:
	//   * accepted values are "yamlpath", "go-yaml", "default" or empty.
	//   * "go-yaml", "default" and empty are equivalent.
	//
	Engine string `yaml:",omitempty"`
	// "file" defines the path of the yaml file to use.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * "file" and "files" are mutually exclusive.
	//   * the schemes "https://", "http://" and "file://" are supported in a source or a condition.
	//
	File string `yaml:",omitempty"`
	// "files" defines the list of yaml file paths to use.
	//
	// compatible:
	//   * condition
	//   * target
	//
	// remark:
	//   * "file" and "files" are mutually exclusive.
	//   * the schemes "https://", "http://" and "file://" are supported in a condition.
	//
	Files []string `yaml:",omitempty"`
	// "key" defines the yaml key path.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// remark:
	//   * "key" is a simpler version of yamlpath.
	//   * in a target, a key that a document does not hold is an error, so that a
	//     manifest never reports success for an update it did not make. The same
	//     rule applies to a wildcard such as `$.agents[*].name`: every position it
	//     selects must hold the key, otherwise the target fails instead of updating
	//     only some of them. Set "searchpattern" to update the positions holding the
	//     key and ignore the others.
	//   * a recursive selector such as `$..name` cannot report a partial match: it
	//     searches for the key itself, so it only selects the positions already
	//     holding it and never fails on the others.
	//   * a field path filtering on a key and value is not supported yet,
	//     see https://github.com/goccy/go-yaml/issues/290
	//
	// example:
	//   * key: $.name
	//   * key: $.agent.name
	//   * key: $.agents[0].name
	//   * key: $.agents[*].name
	//   * key: $.'agents.name'
	//   * key: $.repos[?(@.repository == 'website')].owner (requires engine "yamlpath")
	//
	Key string `yaml:",omitempty"`
	// "keys" defines several yaml key paths to update with the same value.
	//
	// compatible:
	//   * target
	//
	// remark:
	//   * "key" and "keys" are mutually exclusive.
	//   * each entry accepts the same syntax as "key".
	//   * every key is updated with the same value.
	//
	// example:
	//   * keys:
	//     - $.image.tag
	//     - $.sidecar.tag
	//   * keys:
	//     - $.agents[0].version
	//     - $.agents[1].version
	//
	Keys []string `yaml:",omitempty"`
	// "value" defines the value associated with the yaml key.
	//
	// compatible:
	//   * source
	//   * condition
	//   * target
	//
	// default:
	//   in a condition or a target, the output of the associated source.
	//
	Value string `yaml:",omitempty"`
	// "keyonly" checks only that the key exists, whatever its value.
	//
	// compatible:
	//   * condition
	//
	// default:
	//   false
	//
	KeyOnly bool `yaml:",omitempty"`
	// "searchpattern" treats "file" and "files" as path patterns instead of exact paths.
	//
	// The pattern must match the whole path, not just a substring.
	//
	// The pattern syntax is:
	//
	// ```
	//     pattern:
	//         { term }
	//     term:
	//         '*'         matches any sequence of non-Separator characters
	//         '?'         matches any single non-Separator character
	//         '[' [ '^' ] { character-range } ']'
	//                     character class (must be non-empty)
	//         c           matches character c (c != '*', '?', '\\', '[')
	//         '\\' c      matches character c
	//
	//     character-range:
	//         c           matches character c (c != '\\', '-', ']')
	//         '\\' c      matches character c
	//         lo '-' hi   matches character c for lo <= c <= hi
	// ```
	//
	// default:
	//   false
	//
	// remark:
	//   * in a target, it also relaxes the requirement that the key exists: a file
	//     that does not hold it is ignored instead of failing the target, and a
	//     wildcard key updates the positions holding it instead of failing on the
	//     others.
	//
	SearchPattern bool `yaml:",omitempty"`
	// "createmissingkey" creates the key when the yaml document does not hold it yet.
	//
	// compatible:
	//   * target
	//
	// default:
	//   false
	//
	// remark:
	//   * missing intermediate keys are created as nested maps.
	//   * the key is only ever created, never removed, and existing keys are left untouched.
	//   * a missing sequence index such as `$.agents[0].name` cannot be created.
	//   * a key selecting several nodes, such as `$.agents[*].tag` or `$..tag`,
	//     is rejected because the key cannot be created under each selected node.
	//   * not supported by the "yamlpath" engine.
	//   * the yaml file itself must already exist.
	//
	// example:
	//   * key: $.image.tag
	//     createmissingkey: true
	//
	CreateMissingKey bool `yaml:",omitempty"`
	// "appendtoarray" appends the value as a new entry of the yaml sequence selected by "key".
	//
	// compatible:
	//   * target
	//
	// default:
	//   false
	//
	// remark:
	//   * appending is skipped when the sequence already holds the value, so running it twice changes nothing.
	//   * combined with "createmissingkey", a missing sequence is created with the value as its only entry.
	//   * "key" must select the sequence itself, not one of its entries.
	//   * not supported by the "yamlpath" engine.
	//
	// example:
	//   * key: $.allowedTags
	//     appendtoarray: true
	//
	AppendToArray bool `yaml:",omitempty"`
	// "comment" defines a comment added after the value.
	//
	// compatible:
	//   * target
	//
	// default:
	//   empty
	//
	// remark:
	//   * the comment is only added when Updatecli changes the value.
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
	if s.Engine == EngineYamlPath && s.CreateMissingKey {
		validationErrors = append(validationErrors, fmt.Sprintf("Validation error in target of type 'yaml': engine %q does not support the attributes `spec.createmissingkey`", s.Engine))
	}

	if s.Engine == EngineYamlPath && s.AppendToArray {
		validationErrors = append(validationErrors, fmt.Sprintf("Validation error in target of type 'yaml': engine %q does not support the attributes `spec.appendtoarray`", s.Engine))
	}

	// A wildcard or a recursive selector addresses one node per selected position,
	// and there is no way to create a key under each of them: goccy's selector
	// replacement only rewrites the positions that already hold the key, so
	// creating through such a key would silently write nothing.
	if s.CreateMissingKey {
		for _, key := range s.getKeys() {
			// A key that does not parse is reported when it is evaluated, with a
			// message naming the offending character.
			if multiMatch, err := multiMatchKey(key); err == nil && multiMatch {
				validationErrors = append(validationErrors, fmt.Sprintf("Validation error in target of type 'yaml': the attribute `spec.createmissingkey` does not support the wildcard or recursive selector of key %q", key))
			}
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
func (y *Yaml) initFiles(pathResolver pathresolver.Resolver) error {
	y.files = make(map[string]file)

	// File as unique element of newResource.files
	if len(y.spec.File) > 0 {
		var foundFiles []string
		var err error
		switch y.spec.SearchPattern {
		case true:
			foundFiles, err = utils.FindFilesMatchingPathPattern(pathResolver.Dir(), y.spec.File)
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
			foundFiles, err = utils.FindFilesMatchingPathPattern(pathResolver.Dir(), specFile)
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

	return y.UpdateAbsoluteFilePath(pathResolver)
}

// UpdateAbsoluteFilePath resolves every file of the y.files map against the path resolver,
// leaving the path the user wrote available as originalFilePath for reporting.
func (y *Yaml) UpdateAbsoluteFilePath(pathResolver pathresolver.Resolver) error {
	for filePath := range y.files {
		file := y.files[filePath]

		resolvedPath, err := pathResolver.Resolve(file.originalFilePath)
		if err != nil {
			return err
		}

		if resolvedPath != file.filePath {
			logrus.Debugf("Relative path detected: changing from %q to %q", file.originalFilePath, resolvedPath)
		}

		file.filePath = resolvedPath
		y.files[filePath] = file
	}

	return nil
}

// ReportConfig returns a new configuration object with only the necessary fields
// to identify the resource without any sensitive information or context specific data.
func (y *Yaml) ReportConfig() interface{} {
	return Spec{
		File:             y.spec.File,
		Files:            y.spec.Files,
		Key:              y.spec.Key,
		Keys:             y.spec.Keys,
		Value:            y.spec.Value,
		Engine:           y.spec.Engine,
		KeyOnly:          y.spec.KeyOnly,
		SearchPattern:    y.spec.SearchPattern,
		CreateMissingKey: y.spec.CreateMissingKey,
		AppendToArray:    y.spec.AppendToArray,
	}
}

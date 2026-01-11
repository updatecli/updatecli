package yaml

import (
	"errors"
	"fmt"
	"io"
	"strings"

	goyaml "github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/parser"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"

	yamlpath_new "github.com/helm-unittest/yaml-jsonpath/pkg/yamlpath"
	yamlpath_old "github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"go.yaml.in/yaml/v3"
)

// Condition checks if a key exists in a yaml file
func (y *Yaml) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {

	var errorMessages []error

	// Validate information when user want to only check the existence of a YAML key
	if y.spec.KeyOnly && y.spec.Value != "" {
		// Then there must not be any specified Value
		return false, "", fmt.Errorf("validation error in condition of type 'yaml': both `spec.value` and `spec.keyonly` specified while mutually exclusive. Remove one of these 2 directives")
	}

	workDir := ""
	if scm != nil {
		workDir = scm.GetDirectory()
	}

	if err := y.initFiles(workDir); err != nil {
		return false, "", fmt.Errorf("init yaml files: %w", err)
	}

	if len(y.files) == 0 {
		return false, "", fmt.Errorf("no yaml file found")
	}

	// Start by retrieving the specified file's content
	if err := y.Read(); err != nil {
		return false, "", fmt.Errorf("reading yaml file: %w", err)
	}

	// If a source is provided, then the key 'Value' cannot be specified
	valueToCheck := y.spec.Value

	var results []string

	for i := range y.files {
		fileContent := y.files[i].content
		originalFilePath := y.files[i].originalFilePath

		switch y.spec.Engine {
		case EngineGoYaml, EngineDefault, EngineUndefined:
			urlPath, err := goyaml.PathString(y.spec.Key)
			if err != nil {
				errorMessages = append(errorMessages, fmt.Errorf(
					"%q - crafting yamlpath query: %s", originalFilePath, err.Error()))
				continue
			}

			file, err := parser.ParseBytes([]byte(fileContent), 0)
			if err != nil {
				errorMessages = append(errorMessages, fmt.Errorf(
					"%q - parsing yaml file: %s", originalFilePath, err.Error()))
				continue
			}

			for index, doc := range file.Docs {
				if y.spec.DocumentIndex != nil {
					if index != *y.spec.DocumentIndex {
						continue
					}
				}

				node, err := urlPath.FilterNode(doc.Body)
				if err != nil {
					errorMessages = append(errorMessages, fmt.Errorf(
						"%q - searching in yaml file: %w", originalFilePath, err))
					continue
				}

				if node == nil {
					errorMessages = append(errorMessages,
						fmt.Errorf("%q - %w", originalFilePath, ErrKeyNotFound))
					continue
				}

				results = append(results, node.String())
			}

		case EngineYamlPath:
			urlPath, err := yamlpath_old.NewPath(y.spec.Key)
			urlPath1, err := yamlpath_new.NewPath(y.spec.Key)
			if urlPath != urlPath1 {
				fmt.Errorf("%s different from %s", urlPath, urlPath1)
			}
			if err != nil {
				errorMessages = append(errorMessages, fmt.Errorf(
					"%q - crafting yamlpath query: %w", originalFilePath, err))
			}

			// Decode the file into one or more YAML document nodes
			var docs []*yaml.Node
			dec := yaml.NewDecoder(strings.NewReader(y.files[i].content))
			for {
				var doc yaml.Node
				if derr := dec.Decode(&doc); derr != nil {
					if derr == io.EOF {
						break
					}
					errorMessages = append(errorMessages, fmt.Errorf(
						"%q - parsing yaml file: %w", originalFilePath, err))
					continue
				}
				docs = append(docs, &doc)
			}

			for index, doc := range docs {
				if y.spec.DocumentIndex != nil {
					if index != *y.spec.DocumentIndex {
						continue
					}
				}

				founds, err := urlPath.Find(doc)
				if err != nil {

					if err.Error() == "node not found" {
						errorMessages = append(errorMessages, ErrKeyNotFound)
						continue
					}

					errorMessages = append(errorMessages, fmt.Errorf(
						"%q - searching in yaml file: %w", originalFilePath, err))
					continue
				}

				for i := range founds {
					results = append(results, founds[i].Value)
				}
			}

		default:
			return false, "", fmt.Errorf("unsupported yaml engine %q", y.spec.Engine)
		}
	}

	if len(errorMessages) > 0 {
		if y.spec.KeyOnly {
			for i := range errorMessages {
				if !errors.Is(errorMessages[i], ErrKeyNotFound) {

					return false, "", errorsToError(errorMessages)
				}
			}
			return false, "key not found in yaml file(s)", nil
		}

		return false, "", errorsToError(errorMessages)
	}

	originalFilePaths := make([]string, len(y.files))
	for i := range y.files {
		originalFilePaths = append(originalFilePaths, y.files[i].originalFilePath)
	}

	// When user want to only check the existence of a YAML key
	if y.spec.KeyOnly {
		if len(results) == len(y.files) {
			return true, fmt.Sprintf("key %q found in yaml file(s) [%q]", y.spec.Key, strings.Join(originalFilePaths, ",")), nil
		}
		return false, fmt.Sprintf("key %q not found in yaml file(s) [%q]", y.spec.Key, strings.Join(originalFilePaths, ",")), nil
	}

	// When user want to check the value of YAML key and when the input source value is not empty
	if source != "" {
		// Then there must not be any specified Value
		if y.spec.Value != "" {
			return false, "", fmt.Errorf("validation error in condition of type 'yaml': input source value detected, while `spec.value` specified. Add 'disablesourceinput: true' to your manifest to keep ``spec.value`")
		}

		// Use the source input value in this case
		valueToCheck = source
	}

	for _, res := range results {
		if res != valueToCheck {
			return false, fmt.Sprintf("key %q, is incorrectly set to %q and should be %q",
				y.spec.Key,
				res,
				valueToCheck), nil
		}
	}

	return true, fmt.Sprintf("key %q is correctly set to %q", y.spec.Key, valueToCheck), nil
}

func errorsToError(errorMessages []error) error {

	result := []string{}
	if len(errorMessages) > 0 {
		result = append(result, fmt.Sprintf("error detected in condition of type 'yaml': %d error(s) found", len(errorMessages)))
	}

	for i := range errorMessages {
		result = append(result, "\t* "+errorMessages[i].Error())
	}

	return errors.New(strings.Join(result, "\n"))
}

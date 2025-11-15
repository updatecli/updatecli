package yaml

import (
	"errors"
	"fmt"
	"os"

	goyaml "github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/parser"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"

	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
)

// Source return the latest version
func (y *Yaml) Source(workingDir string, resultSource *result.Source) error {
	// By default workingDir is set to local directory
	var filePath string

	// By the default workingdir is set to the current working directory
	// it would be better to have it empty by default but it must be changed in the
	// source core codebase.
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		return errors.New("fail getting current working directory")
	}

	if y.spec.SearchPattern {
		return fmt.Errorf("validation error in sources of type 'yaml': the attribute `spec.searchpattern` is not supported for source")
	}

	if len(y.files) > 1 {
		validationError := fmt.Errorf("validation error in sources of type 'yaml': the attributes `spec.files` can't contain more than one element for sources")
		logrus.Errorln(validationError.Error())
		return validationError
	}

	if y.spec.Value != "" {
		logrus.Warnf("Key 'Value' is not used by source YAML")
	}

	if workingDir == currentWorkingDirectory {
		workingDir = ""
	}

	if err := y.initFiles(workingDir); err != nil {
		return fmt.Errorf("init files: %w", err)
	}

	switch len(y.files) {
	case 1:
		//
	case 0:
		return fmt.Errorf("no yaml file found")
	default:
		return fmt.Errorf("multiple yaml files found, please specify only one file")
	}

	// loop over the only file
	for f := range y.files {
		filePath = f

		// Ideally currentWorkingDirectory should be empty
		if workingDir != currentWorkingDirectory {
			y.UpdateAbsoluteFilePath(workingDir)
		}
	}

	if err = y.Read(); err != nil {
		return fmt.Errorf("reading yaml file: %w", err)
	}

	fileContent := y.files[filePath].content
	originalFilePath := y.files[filePath].originalFilePath

	var results []string
	switch y.spec.Engine {
	case EngineGoYaml, EngineDefault, EngineUndefined:
		urlPath, err := goyaml.PathString(y.spec.Key)
		if err != nil {
			return fmt.Errorf("crafting yamlpath query: %w", err)
		}

		file, err := parser.ParseBytes([]byte(fileContent), 0)
		if err != nil {
			return fmt.Errorf("parsing yaml file: %w", err)
		}

		switch y.spec.DocumentIndex {
		case nil:
			node, err := urlPath.FilterFile(file)
			if err != nil && !errors.Is(err, goyaml.ErrNotFoundNode) {
				return fmt.Errorf("searching in yaml file: %w", err)
			}

			if node != nil {
				results = append(results, node.String())
			}

		default:
			for index, doc := range file.Docs {

				if index != *y.spec.DocumentIndex {
					continue
				}

				node, err := urlPath.FilterNode(doc.Body)
				if err != nil {
					return fmt.Errorf("searching in yaml document index %d: %w", *y.spec.DocumentIndex, err)
				}

				if node != nil {
					results = append(results, node.String())
					break
				}
			}
		}

	case EngineYamlPath:
		urlPath, err := yamlpath.NewPath(y.spec.Key)
		if err != nil {
			return fmt.Errorf("crafting yamlpath query: %w", err)
		}

		var n yaml.Node
		err = yaml.Unmarshal([]byte(fileContent), &n)
		if err != nil {
			return fmt.Errorf("parsing yaml file: %w", err)
		}

		founds, err := urlPath.Find(&n)
		if err != nil {
			return fmt.Errorf("searching in yaml file: %w", err)
		}

		for i := range founds {
			results = append(results, founds[i].Value)
		}

	default:
		return fmt.Errorf("unsupported engine %q", y.spec.Engine)
	}
	if len(results) == 0 {
		return fmt.Errorf("impossible to find any key for the specified path: %s", y.spec.Key)
	}

	if len(results) > 0 {
		value := results[0]
		resultSource.Result = result.SUCCESS
		resultSource.Information = value
		resultSource.Description = fmt.Sprintf("value %q found for key %q in the yaml file %q",
			value,
			y.spec.Key,
			originalFilePath,
		)
		return nil
	}

	return fmt.Errorf("cannot find key %q from file %q",
		y.spec.Key,
		originalFilePath)
}

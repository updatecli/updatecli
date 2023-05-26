package yaml

import (
	"errors"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/parser"
)

// Condition checks if a key exists in a yaml file
func (y *Yaml) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {
	var fileContent string
	var originalFilePath string

	if scm != nil {
		y.UpdateAbsoluteFilePath(scm.GetDirectory())
	}

	// Start by retrieving the specified file's content
	if err := y.Read(); err != nil {
		return fmt.Errorf("reading yaml file: %w", err)
	}

	// loop over the only file
	for theFilePath := range y.files {
		fileContent = y.files[theFilePath].content
		originalFilePath = y.files[theFilePath].originalFilePath
	}

	// If a source is provided, then the key 'Value' cannot be specified
	valueToCheck := y.spec.Value

	urlPath, err := yaml.PathString(y.spec.Key)
	if err != nil {
		return fmt.Errorf("crafting yamlpath query: %w", err)
	}

	file, err := parser.ParseBytes([]byte(fileContent), 0)
	if err != nil {
		return fmt.Errorf("parsing yaml file: %w", err)
	}

	node, err := urlPath.FilterFile(file)
	if err != nil && !errors.Is(err, yaml.ErrNotFoundNode) {
		return fmt.Errorf("searching in yaml file: %w", err)
	}

	// When user want to only check the existence of a YAML key
	if y.spec.KeyOnly {
		// Then there must not be any specified Value
		if y.spec.Value != "" {
			validationError := fmt.Errorf("validation error in condition of type 'yaml': both `spec.value` and `spec.keyonly` specified while mutually exclusive. Remove one of these 2 directives")
			return validationError
		}

		if node != nil {
			resultCondition.Result = result.SUCCESS
			resultCondition.Pass = true
			resultCondition.Description = fmt.Sprintf("key %q found in yaml file %q", y.spec.Key, y.spec.File)
			return nil
		}

		resultCondition.Result = result.FAILURE
		resultCondition.Pass = false
		resultCondition.Description = fmt.Sprintf("key %q not found in yaml file %q", y.spec.Key, y.spec.File)

		return nil
	}

	// When user want to check the value of YAML key and when the input source value is not empty
	if source != "" {
		// Then there must not be any specified Value
		if y.spec.Value != "" {
			validationError := fmt.Errorf("validation error in condition of type 'yaml': input source value detected, while `spec.value` specified. Add 'disablesourceinput: true' to your manifest to keep ``spec.value`")
			return validationError
		}

		// Use the source input value in this case
		valueToCheck = source
	}

	if node != nil {
		if node.String() == valueToCheck {
			resultCondition.Description = fmt.Sprintf("key %q, in YAML file %q, is correctly set to %q",
				y.spec.Key,
				originalFilePath,
				valueToCheck,
			)

			resultCondition.Pass = true
			resultCondition.Result = result.SUCCESS

			return nil
		}

		resultCondition.Pass = false
		resultCondition.Result = result.FAILURE
		resultCondition.Description = fmt.Sprintf("key %q, in YAML file %q, is incorrectly set to %q and should be %q",
			y.spec.Key,
			originalFilePath,
			node.String(),
			valueToCheck)
		return nil
	}

	return fmt.Errorf("%s cannot find key %q in the YAML file %q",
		result.FAILURE,
		y.spec.Key,
		originalFilePath,
	)
}

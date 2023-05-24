package yaml

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/parser"
)

// https://github.com/goccy/go-yaml/issues/217

// Target updates a scm repository based on the modified yaml file.
func (y *Yaml) Target(source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {

	if scm != nil {
		y.UpdateAbsoluteFilePath(scm.GetDirectory())
	}

	// Test if target reference a file with a prefix like https:// or file://, as we don't know how to update those files.
	for _, file := range y.files {
		if text.IsURL(file.originalFilePath) {
			return fmt.Errorf("unsupported filename prefix for %s", file.originalFilePath)
		}

		if strings.HasPrefix(file.originalFilePath, "https://") ||
			strings.HasPrefix(file.originalFilePath, "http://") {
			return fmt.Errorf("URL scheme is not supported for YAML target: %q", file.originalFilePath)
		}

		// Test at runtime if a file exist (no ForceCreate for kind: yaml)
		if !y.contentRetriever.FileExists(file.filePath) {
			return fmt.Errorf("the yaml file %q does not exist", file.originalFilePath)
		}
	}

	if err := y.Read(); err != nil {
		return fmt.Errorf("loading yaml file(s): %w", err)
	}

	valueToWrite := source
	if y.spec.Value != "" {
		valueToWrite = y.spec.Value
		logrus.Info("INFO: Using spec.Value instead of source input value.")
	}

	resultTarget.NewInformation = valueToWrite

	shouldMsg := " "
	if dryRun {
		// Use to craft message depending if we run Updatecli in dryrun mode or not
		shouldMsg = " should be "
	}

	// loop over file(s)
	notChanged := 0
	//originalContents := make(map[string]string)

	urlPath, err := yaml.PathString(y.spec.Key)
	if err != nil {
		return fmt.Errorf("crafting yamlpath query: %w", err)
	}

	for filePath := range y.files {
		originFilePath := y.files[filePath].originalFilePath

		yamlFile, err := parser.ParseBytes([]byte(y.files[filePath].content), parser.ParseComments)
		if err != nil {
			return fmt.Errorf("parsing yaml file: %w", err)
		}

		node, err := urlPath.FilterFile(yamlFile)
		if err != nil {
			if errors.Is(err, yaml.ErrNotFoundNode) {
				return fmt.Errorf("couldn't find key %q from file %q",
					y.spec.Key,
					originFilePath)
			}
			return fmt.Errorf("searching in yaml file: %w", err)
		}

		oldVersion := node.String()
		resultTarget.OldInformation = oldVersion

		if oldVersion == valueToWrite {
			resultTarget.Description = fmt.Sprintf("%s\nkey %q already set to %q, from file %q, ",
				resultTarget.Description,
				y.spec.Key,
				valueToWrite,
				originFilePath)
			notChanged++
			continue
		}

		if err := urlPath.ReplaceWithReader(yamlFile, strings.NewReader(valueToWrite)); err != nil {
			return fmt.Errorf("replacing yaml key: %w", err)
		}

		f := y.files[filePath]
		f.content = yamlFile.String()
		y.files[filePath] = f

		resultTarget.Changed = true
		resultTarget.Files = append(resultTarget.Files, y.files[filePath].filePath)
		resultTarget.Result = result.ATTENTION

		resultTarget.Description = fmt.Sprintf("%s\nkey %q%supdated from %q to %q, in file %q",
			resultTarget.Description,
			y.spec.Key,
			shouldMsg,
			oldVersion,
			valueToWrite,
			originFilePath)

		if !dryRun {
			newFile, err := os.Create(y.files[filePath].filePath)

			// https://staticcheck.io/docs/checks/#SA5001
			//lint:ignore SA5001 We want to defer the file closing before exiting the function
			defer newFile.Close()

			if err != nil {
				return fmt.Errorf("creating file %q: %w", originFilePath, err)
			}

			err = y.contentRetriever.WriteToFile(
				y.files[filePath].content,
				y.files[filePath].filePath)

			if err != nil {
				return fmt.Errorf("saving file %q: %w", originFilePath, err)
			}
		}
	}

	resultTarget.Description = strings.TrimPrefix(resultTarget.Description, "\n")

	// If no file was updated, don't return an error
	if notChanged == len(y.files) {
		resultTarget.Result = result.SUCCESS
		return nil
	}

	sort.Strings(resultTarget.Files)

	return nil
}

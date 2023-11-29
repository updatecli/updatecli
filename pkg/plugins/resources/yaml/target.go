package yaml

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"

	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"

	goyaml "github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/parser"
)

// Target updates a scm repository based on the modified yaml file.
func (y *Yaml) Target(source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {
	var err error

	workDir := ""
	if scm != nil {
		workDir = scm.GetDirectory()
	}

	if err := y.initFiles(workDir); err != nil {
		return fmt.Errorf("init files: %w", err)
	}

	if len(y.files) == 0 {
		return fmt.Errorf("no yaml file found")
	}

	// Test if target reference a file with a prefix like https:// or file://, as we don't know how to update those files.
	if err = y.validateTargetFilePath(); err != nil {
		return fmt.Errorf("filepath validation error: %w", err)
	}

	if err = y.Read(); err != nil {
		return fmt.Errorf("loading yaml file(s): %w", err)
	}

	valueToWrite := source
	if y.spec.Value != "" {
		valueToWrite = y.spec.Value
		logrus.Debug("Using spec.Value instead of source input value.")
	}

	resultTarget.NewInformation = valueToWrite

	var notChanged int
	switch y.spec.Engine {
	case EngineGoYaml, EngineDefault, EngineUndefined:
		notChanged, err = y.goYamlTarget(valueToWrite, resultTarget, dryRun)
		if err != nil {
			return fmt.Errorf("updating yaml file: %w", err)
		}

	case EngineYamlPath:
		notChanged, err = y.goYamlPathTarget(valueToWrite, resultTarget, dryRun)
		if err != nil {
			return fmt.Errorf("updating yaml file: %w", err)
		}

	default:
		return fmt.Errorf("unsupported engine %q", y.spec.Engine)
	}

	resultTarget.Description = strings.TrimPrefix(resultTarget.Description, "\n")

	// If no file was updated, don't return an error
	if notChanged == len(y.files) {
		resultTarget.Changed = false
		resultTarget.Result = result.SUCCESS
		return nil
	}

	sort.Strings(resultTarget.Files)

	return nil
}

func (y Yaml) validateTargetFilePath() error {
	var errs []error
	for _, file := range y.files {
		if text.IsURL(file.originalFilePath) {
			errs = append(errs, fmt.Errorf("%s: unsupported filename prefix", file.originalFilePath))
		}

		if strings.HasPrefix(file.originalFilePath, "https://") ||
			strings.HasPrefix(file.originalFilePath, "http://") {
			errs = append(errs, fmt.Errorf("%s: URL scheme is not supported for YAML target", file.originalFilePath))
		}

		// Test at runtime if a file exist (no ForceCreate for kind: yaml)
		if !y.contentRetriever.FileExists(file.filePath) {
			errs = append(errs, fmt.Errorf("%s: the yaml file does not exist", file.originalFilePath))
		}
	}

	if len(errs) > 0 {
		for i := range errs {
			logrus.Errorln(errs[i])
		}
		return fmt.Errorf("invalid target file path")
	}
	return nil
}

func (y Yaml) goYamlTarget(valueToWrite string, resultTarget *result.Target, dryRun bool) (notChanged int, err error) {

	urlPath, err := goyaml.PathString(y.spec.Key)
	if err != nil {
		return 0, fmt.Errorf("crafting yamlpath query: %w", err)
	}

	for filePath := range y.files {
		originFilePath := y.files[filePath].originalFilePath

		yamlFile, err := parser.ParseBytes([]byte(y.files[filePath].content), parser.ParseComments)
		if err != nil {
			return 0, fmt.Errorf("parsing yaml file: %w", err)
		}

		node, err := urlPath.FilterFile(yamlFile)
		if err != nil {
			if y.spec.SearchPattern {
				// If search pattern is true then we don't want to return an error
				// as we are probably trying to identify a file matching the key
				logrus.Debugf("ignoring file %q as we couldn't find key %q", originFilePath, y.spec.Key)
				continue
			}
			if errors.Is(err, goyaml.ErrNotFoundNode) {
				return 0, fmt.Errorf("couldn't find key %q from file %q",
					y.spec.Key,
					originFilePath)
			}
			return 0, fmt.Errorf("searching in yaml file: %w", err)
		}

		oldVersion := node.String()
		resultTarget.Information = oldVersion

		if oldVersion == valueToWrite {
			resultTarget.Description = fmt.Sprintf("%s\nkey %q already set to %q, from file %q",
				resultTarget.Description,
				y.spec.Key,
				valueToWrite,
				originFilePath)
			notChanged++
			continue
		}

		if err := urlPath.ReplaceWithReader(yamlFile, strings.NewReader(valueToWrite)); err != nil {
			return 0, fmt.Errorf("replacing yaml key: %w", err)
		}

		f := y.files[filePath]
		f.content = yamlFile.String()
		y.files[filePath] = f

		resultTarget.Changed = true
		resultTarget.Files = append(resultTarget.Files, y.files[filePath].filePath)
		resultTarget.Result = result.ATTENTION

		shouldMsg := " "
		if dryRun {
			// Use to craft message depending if we run Updatecli in dryrun mode or not
			shouldMsg = " should be "
		}

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
				return 0, fmt.Errorf("creating file %q: %w", originFilePath, err)
			}

			err = y.contentRetriever.WriteToFile(
				y.files[filePath].content,
				y.files[filePath].filePath)

			if err != nil {
				return 0, fmt.Errorf("saving file %q: %w", originFilePath, err)
			}
		}
	}

	return notChanged, nil
}

func (y *Yaml) goYamlPathTarget(valueToWrite string, resultTarget *result.Target, dryRun bool) (notChanged int, err error) {
	urlPath, err := yamlpath.NewPath(y.spec.Key)
	if err != nil {
		return 0, fmt.Errorf("crafting yamlpath query: %w", err)
	}

	var buf bytes.Buffer
	e := yaml.NewEncoder(&buf)
	defer e.Close()
	e.SetIndent(2)

	for filePath := range y.files {
		buf = bytes.Buffer{}
		originFilePath := y.files[filePath].originalFilePath

		var n yaml.Node

		err = yaml.Unmarshal([]byte(y.files[filePath].content), &n)
		if err != nil {
			return 0, fmt.Errorf("parsing yaml file: %w", err)
		}

		nodes, err := urlPath.Find(&n)
		if err != nil {
			return 0, fmt.Errorf("searching in yaml file: %w", err)
		}

		if len(nodes) == 0 {
			if y.spec.SearchPattern {
				// If search pattern is true then we don't want to return an error
				// as we are probably trying to identify a file matching the key
				logrus.Debugf("ignoring file %q as we couldn't find key %q", originFilePath, y.spec.Key)
				continue
			}
			return 0, fmt.Errorf("couldn't find key %q from file %q",
				y.spec.Key,
				originFilePath)
		}

		var oldVersion string
		for _, node := range nodes {
			oldVersion = node.Value
			resultTarget.Information = oldVersion

			if oldVersion == valueToWrite {
				resultTarget.Description = fmt.Sprintf("%s\nkey %q already set to %q, from file %q, ",
					resultTarget.Description,
					y.spec.Key,
					valueToWrite,
					originFilePath)
				notChanged++
				continue
			}

			node.Value = valueToWrite
		}

		f := y.files[filePath]
		err = e.Encode(&n)
		if err != nil {
			return 0, fmt.Errorf("unable to marshal the yaml file: %w", err)
		}

		//
		f.content = buf.String()
		if strings.HasPrefix(y.files[filePath].content, "---\n") &&
			!strings.HasPrefix(f.content, "---\n") {
			f.content = "---\n" + f.content
		}
		y.files[filePath] = f

		resultTarget.Changed = true
		resultTarget.Files = append(resultTarget.Files, y.files[filePath].filePath)
		resultTarget.Result = result.ATTENTION

		shouldMsg := " "
		if dryRun {
			// Use to craft message depending if we run Updatecli in dryrun mode or not
			shouldMsg = " should be "
		}

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
				return 0, fmt.Errorf("creating file %q: %w", originFilePath, err)
			}

			err = y.contentRetriever.WriteToFile(
				y.files[filePath].content,
				y.files[filePath].filePath)

			if err != nil {
				return 0, fmt.Errorf("saving file %q: %w", originFilePath, err)
			}
		}
	}
	return notChanged, nil
}

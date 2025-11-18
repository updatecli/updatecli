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
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/goccy/go-yaml/token"
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

	// notChanged is used to count the number of files that have not been updated
	var notChanged int
	// ignoredFiles is used to count the number of files that have been ignored for example because the key was not found and searchPattern is true
	var ignoredFiles int

	switch y.spec.Engine {
	case EngineGoYaml, EngineDefault, EngineUndefined:
		notChanged, ignoredFiles, err = y.goYamlTarget(valueToWrite, resultTarget, dryRun)
		if err != nil {
			return fmt.Errorf("updating yaml file: %w", err)
		}

	case EngineYamlPath:
		notChanged, ignoredFiles, err = y.goYamlPathTarget(valueToWrite, resultTarget, dryRun)
		if err != nil {
			return fmt.Errorf("updating yaml file: %w", err)
		}

	default:
		return fmt.Errorf("unsupported engine %q", y.spec.Engine)
	}

	resultTarget.Description = strings.TrimPrefix(resultTarget.Description, "\n")

	monitoredFiles := len(y.files) - ignoredFiles

	if monitoredFiles == 0 {
		resultTarget.Description = "no yaml file found matching criteria"
		resultTarget.Result = result.SKIPPED
		return nil
	}

	// If no file was updated, don't return an error
	if notChanged == monitoredFiles && monitoredFiles > 0 {
		resultTarget.Description = fmt.Sprintf("no change detected:\n\t* %s", strings.ReplaceAll(strings.TrimPrefix(resultTarget.Description, "\n"), "\n", "\n\t* "))
		resultTarget.Changed = false
		resultTarget.Result = result.SUCCESS
		return nil
	}

	resultTarget.Description = fmt.Sprintf("change detected:\n\t* %s", strings.ReplaceAll(strings.TrimPrefix(resultTarget.Description, "\n"), "\n", "\n\t* "))

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

func (y Yaml) goYamlTarget(valueToWrite string, resultTarget *result.Target, dryRun bool) (notChanged int, ignoredFiles int, err error) {

	nodeToWrite, err := goyaml.ValueToNode(valueToWrite)

	if y.spec.Comment != "" {
		commentGroup := &ast.CommentGroupNode{
			Comments: []*ast.CommentNode{
				{
					Token: &token.Token{
						Type: token.CommentType,
						// Add a space before the comment
						Value: " " + y.spec.Comment,
					},
				},
			},
		}

		err = nodeToWrite.SetComment(commentGroup)
		if err != nil {
			logrus.Errorf("error setting comment: %s", err)
		}
	}

	if err != nil {
		return 0, ignoredFiles, fmt.Errorf("parsing value to write: %w", err)
	}

	keys := y.spec.getKeys()

	resultTargetFilesMap := map[string]bool{}

	for filePath := range y.files {
		originFilePath := y.files[filePath].originalFilePath
		fileNotChanged := 0
		fileKeysProcessed := 0

		yamlFile, err := parser.ParseBytes([]byte(y.files[filePath].content), parser.ParseComments)
		if err != nil {
			return 0, ignoredFiles, fmt.Errorf("parsing yaml file: %w", err)
		}

		// Process each key for this file
		for _, key := range keys {
			urlPath, err := goyaml.PathString(key)
			if err != nil {
				return 0, ignoredFiles, fmt.Errorf("crafting yamlpath query for key %q: %w", key, err)
			}

			oldVersion := ""
			keyNotFound := []string{}
			errMsg := []string{}
			contentChanged := false

			for index, doc := range yamlFile.Docs {
				var node ast.Node

				if y.spec.DocumentIndex != nil {
					if index != *y.spec.DocumentIndex {
						continue
					}
				}

				node, err = urlPath.FilterNode(doc.Body)
				if err != nil {
					if errors.Is(err, goyaml.ErrNotFoundNode) {
						if y.spec.SearchPattern {
							// If search pattern is true then we don't want to return an error
							// as we are probably trying to identify a file matching the key
							logrus.Debugf("ignoring key %q from file %q in document %q: %s", key, originFilePath, index, err)
							continue
						}
						keyNotFound = append(keyNotFound, fmt.Sprintf("couldn't find key %q from file %q in document %q", key, originFilePath, index))
					}

					errMsg = append(errMsg, fmt.Sprintf("searching for key %q in document index %d: %s", key, index, err.Error()))
					continue
				}

				if node == nil {
					keyNotFound = append(keyNotFound, fmt.Sprintf("couldn't find key %q from file %q", key, originFilePath))
					continue
				}

				oldVersion = node.String()
				if node.String() != nodeToWrite.String() && node.String() != valueToWrite {
					contentChanged = true
				}
			}

			if len(errMsg) > 0 {
				return 0, ignoredFiles, fmt.Errorf("errors occurred:\n%s", strings.Join(errMsg, "\n"))
			}

			if len(keyNotFound) > 0 {
				for _, msg := range keyNotFound {
					logrus.Errorln(msg)
				}
				return 0, ignoredFiles, fmt.Errorf("key not found from file %q", originFilePath)
			}

			fileKeysProcessed++
			resultTarget.Information = oldVersion

			if !contentChanged {
				resultTarget.Description = fmt.Sprintf("%s\nkey %q already set to %q, from file %q",
					resultTarget.Description,
					key,
					valueToWrite,
					originFilePath)
				fileNotChanged++
				continue
			}

			for index, doc := range yamlFile.Docs {
				if y.spec.DocumentIndex != nil {
					if index != *y.spec.DocumentIndex {
						continue
					}
				}

				tmpYAMLFile := ast.File{
					Name: yamlFile.Name,
				}
				tmpYAMLFile.Docs = append(tmpYAMLFile.Docs, doc)
				if err := urlPath.ReplaceWithNode(&tmpYAMLFile, nodeToWrite); err != nil {
					return 0, ignoredFiles, fmt.Errorf("replacing yaml key %q: %w", key, err)
				}
				yamlFile.Docs[index].Body = tmpYAMLFile.Docs[0].Body
			}

			if _, ok := resultTargetFilesMap[filePath]; !ok {
				resultTarget.Files = append(resultTarget.Files, y.files[filePath].filePath)
				resultTargetFilesMap[filePath] = true
			}

			resultTarget.Changed = true
			resultTarget.Result = result.ATTENTION

			shouldMsg := " "
			if dryRun {
				// Use to craft message depending if we run Updatecli in dryrun mode or not
				shouldMsg = " should be "
			}

			resultTarget.Description = fmt.Sprintf("%s\nkey %q%supdated from %q to %q, in file %q",
				resultTarget.Description,
				key,
				shouldMsg,
				oldVersion,
				valueToWrite,
				originFilePath)
		}

		// If no keys were processed for this file (all were ignored), count as ignored
		if fileKeysProcessed == 0 {
			ignoredFiles++
			continue
		}

		// If all processed keys in this file were unchanged, count as not changed
		if fileNotChanged == fileKeysProcessed {
			notChanged++
		}

		// Update file content
		f := y.files[filePath]
		f.content = yamlFile.String()
		y.files[filePath] = f

		if !dryRun {
			newFile, err := os.Create(y.files[filePath].filePath)
			if err != nil {
				return 0, ignoredFiles, fmt.Errorf("creating file %q: %w", originFilePath, err)
			}
			defer newFile.Close()

			err = y.contentRetriever.WriteToFile(
				y.files[filePath].content,
				y.files[filePath].filePath)

			if err != nil {
				return 0, ignoredFiles, fmt.Errorf("saving file %q: %w", originFilePath, err)
			}
		}
	}

	return notChanged, ignoredFiles, nil
}

func (y *Yaml) goYamlPathTarget(valueToWrite string, resultTarget *result.Target, dryRun bool) (notChanged int, ignoredFiles int, err error) {
	var buf bytes.Buffer
	e := yaml.NewEncoder(&buf)
	defer e.Close()

	e.SetIndent(2)

	keys := y.spec.getKeys()

	resultTargetFilesMap := map[string]bool{}

	for filePath := range y.files {
		buf = bytes.Buffer{}
		originFilePath := y.files[filePath].originalFilePath
		fileNotChanged := 0
		fileKeysProcessed := 0

		var n yaml.Node

		err = yaml.Unmarshal([]byte(y.files[filePath].content), &n)
		if err != nil {
			return 0, ignoredFiles, fmt.Errorf("parsing yaml file: %w", err)
		}

		// Process each key for this file
		for _, key := range keys {
			urlPath, err := yamlpath.NewPath(key)
			if err != nil {
				return 0, 0, fmt.Errorf("crafting yamlpath query for key %q: %w", key, err)
			}

			nodes, err := urlPath.Find(&n)
			if err != nil {
				return 0, ignoredFiles, fmt.Errorf("searching for key %q in yaml file: %w", key, err)
			}

			if len(nodes) == 0 {
				if y.spec.SearchPattern {
					// If search pattern is true then we don't want to return an error
					// as we are probably trying to identify a file matching the key
					logrus.Debugf("ignoring key %q in file %q as we couldn't find it", key, originFilePath)
					continue
				}
				return 0, ignoredFiles, fmt.Errorf("couldn't find key %q from file %q",
					key,
					originFilePath)
			}

			fileKeysProcessed++
			var oldVersion string
			var notChangedNode int
			for _, node := range nodes {
				oldVersion = node.Value
				resultTarget.Information = oldVersion

				if oldVersion == valueToWrite {
					resultTarget.Description = fmt.Sprintf("%s\nkey %q already set to %q, from file %q, ",
						resultTarget.Description,
						key,
						valueToWrite,
						originFilePath)
					notChangedNode++
					continue
				}

				node.Value = valueToWrite
				if y.spec.Comment != "" {
					node.LineComment = y.spec.Comment
				}

				if _, ok := resultTargetFilesMap[filePath]; !ok {
					resultTarget.Files = append(resultTarget.Files, y.files[filePath].filePath)
					resultTargetFilesMap[filePath] = true
				}

				resultTarget.Changed = true
				resultTarget.Result = result.ATTENTION

				shouldMsg := " "
				if dryRun {
					// Use to craft message depending if we run Updatecli in dryrun mode or not
					shouldMsg = " should be "
				}

				resultTarget.Description = fmt.Sprintf("%s\nkey %q%supdated from %q to %q, in file %q",
					resultTarget.Description,
					key,
					shouldMsg,
					oldVersion,
					valueToWrite,
					originFilePath)
			}

			if notChangedNode == len(nodes) {
				fileNotChanged++
			}
		}

		// If no keys were processed for this file (all were ignored), count as ignored
		if fileKeysProcessed == 0 {
			ignoredFiles++
			continue
		}

		// If all processed keys in this file were unchanged, count as not changed
		if fileNotChanged == fileKeysProcessed {
			notChanged++
		}

		// Update file content
		f := y.files[filePath]
		err = e.Encode(&n)
		if err != nil {
			return 0, ignoredFiles, fmt.Errorf("unable to marshal the yaml file: %w", err)
		}

		f.content = buf.String()
		if strings.HasPrefix(y.files[filePath].content, "---\n") &&
			!strings.HasPrefix(f.content, "---\n") {
			f.content = "---\n" + f.content
		}
		y.files[filePath] = f

		if !dryRun {
			newFile, err := os.Create(y.files[filePath].filePath)
			if err != nil {
				return 0, ignoredFiles, fmt.Errorf("creating file %q: %w", originFilePath, err)
			}
			defer newFile.Close()

			err = y.contentRetriever.WriteToFile(
				y.files[filePath].content,
				y.files[filePath].filePath)

			if err != nil {
				return 0, ignoredFiles, fmt.Errorf("saving file %q: %w", originFilePath, err)
			}
		}
	}
	return notChanged, ignoredFiles, nil
}

package yaml

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/result"

	goyaml "github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

func (y Yaml) goYamlTarget(valueToWrite string, resultTarget *result.Target, dryRun bool) (notChanged int, ignoredFiles int, err error) {
	nodeToWrite, err := goyaml.ValueToNode(valueToWrite)
	if err != nil {
		return 0, ignoredFiles, fmt.Errorf("parsing value to write: %w", err)
	}

	if y.spec.Comment != "" {
		if err := setNodeComment(nodeToWrite, y.spec.Comment); err != nil {
			logrus.Errorf("error setting comment: %s", err)
		}
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

			keyNotFound := []string{}
			errMsg := []string{}
			// keyProcessed reports whether the key was resolved, or created, in at
			// least one of the evaluated documents.
			keyProcessed := false
			keyChanged := false

			for index, doc := range yamlFile.Docs {
				if y.spec.DocumentIndex != nil {
					if index != *y.spec.DocumentIndex {
						continue
					}
				}

				// goccy reports a missing key as a nil node rather than an error,
				// but reports an intermediate null value ("key:") as an invalid
				// query. Both mean the key is absent, so when we are allowed to
				// create it we let our own walker decide rather than FilterNode.
				node, filterErr := urlPath.FilterNode(doc.Body)

				switch {
				// A null value is an existing node that holds nothing, so when we
				// may create the key we overwrite it instead of updating it.
				case filterErr == nil && node != nil && (!y.spec.CreateMissingKey || !isNullNode(node)):
					changed, err := y.updateNode(yamlFile, index, doc, urlPath, node, key, valueToWrite, nodeToWrite, resultTarget, originFilePath, dryRun)
					if err != nil {
						return 0, ignoredFiles, err
					}
					keyProcessed = true
					keyChanged = keyChanged || changed

				case y.spec.CreateMissingKey:
					changed, err := y.createKey(doc, key, valueToWrite, resultTarget, originFilePath, dryRun)
					if err != nil {
						return 0, ignoredFiles, err
					}
					keyProcessed = true
					keyChanged = keyChanged || changed

				case filterErr != nil:
					errMsg = append(errMsg, fmt.Sprintf("searching for key %q in document index %d: %s", key, index, filterErr.Error()))

				case y.spec.SearchPattern:
					// If search pattern is true then we don't want to return an error
					// as we are probably trying to identify a file matching the key
					logrus.Debugf("ignoring key %q from file %q in document %d as it does not exist", key, originFilePath, index)

				default:
					keyNotFound = append(keyNotFound, fmt.Sprintf("couldn't find key %q from file %q in document %d", key, originFilePath, index))
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

			if !keyProcessed {
				continue
			}

			fileKeysProcessed++

			if !keyChanged {
				fileNotChanged++
				continue
			}

			if _, ok := resultTargetFilesMap[filePath]; !ok {
				resultTarget.Files = append(resultTarget.Files, y.files[filePath].filePath)
				resultTargetFilesMap[filePath] = true
			}

			resultTarget.Changed = true
			resultTarget.Result = result.ATTENTION
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

// createKey inserts a key that does not exist yet in doc. When the target also
// appends to an array, the key is created holding the value as its sole entry.
func (y Yaml) createKey(doc *ast.DocumentNode, key, valueToWrite string, resultTarget *result.Target, originFilePath string, dryRun bool) (bool, error) {
	elements, err := splitYamlPathKey(key)
	if err != nil {
		return false, fmt.Errorf("cannot create key %q: %w", key, err)
	}

	var leafValue interface{} = valueToWrite
	if y.spec.AppendToArray {
		leafValue = []interface{}{valueToWrite}
	}

	leaf, err := createMissingKey(doc, key, elements, leafValue)
	if err != nil {
		return false, err
	}

	if y.spec.Comment != "" {
		if y.spec.AppendToArray {
			// The comment belongs to the sole entry of the created sequence.
			if sequence, ok := leaf.(*ast.SequenceNode); ok && len(sequence.Values) == 1 {
				leaf = sequence.Values[0]
			}
		}
		if err := setNodeComment(leaf, y.spec.Comment); err != nil {
			logrus.Errorf("error setting comment: %s", err)
		}
	}

	resultTarget.Description = fmt.Sprintf("%s\nkey %q%screated in file %q with value %q",
		resultTarget.Description,
		key,
		shouldMessage(dryRun),
		originFilePath,
		valueToWrite)

	return true, nil
}

// updateNode either appends to the sequence addressed by key, or replaces the value
// of the node it addresses. It reports whether the document was modified.
func (y Yaml) updateNode(yamlFile *ast.File, index int, doc *ast.DocumentNode, urlPath *goyaml.Path, node ast.Node, key, valueToWrite string, nodeToWrite ast.Node, resultTarget *result.Target, originFilePath string, dryRun bool) (bool, error) {
	if y.spec.AppendToArray {
		// A key matching several nodes resolves to a detached wrapper, so append to
		// each matched sequence rather than to the node goccy returned.
		sequences, err := appendTargetSequences(node, key, originFilePath)
		if err != nil {
			return false, err
		}

		appended := false
		for _, sequence := range sequences {
			sequenceAppended, err := appendToSequence(sequence, valueToWrite, y.spec.Comment)
			if err != nil {
				return false, fmt.Errorf("appending to key %q from file %q: %w", key, originFilePath, err)
			}
			appended = appended || sequenceAppended
		}

		if !appended {
			resultTarget.Description = fmt.Sprintf("%s\nkey %q already contains %q, from file %q",
				resultTarget.Description,
				key,
				valueToWrite,
				originFilePath)
			return false, nil
		}

		resultTarget.Description = fmt.Sprintf("%s\nvalue %q%sappended to key %q, in file %q",
			resultTarget.Description,
			valueToWrite,
			shouldMessage(dryRun),
			key,
			originFilePath)

		return true, nil
	}

	oldVersion := node.String()
	resultTarget.Information = oldVersion

	// Compare decoded value so folded/literal scalars (>-, |) aren't
	// flagged as changed by their formatting markers. See issue #8295.
	var decoded string
	if err := goyaml.NodeToValue(node, &decoded); err == nil && decoded == valueToWrite {
		resultTarget.Description = fmt.Sprintf("%s\nkey %q already set to %q, from file %q",
			resultTarget.Description,
			key,
			valueToWrite,
			originFilePath)
		return false, nil
	}

	tmpYAMLFile := ast.File{
		Name: yamlFile.Name,
	}
	tmpYAMLFile.Docs = append(tmpYAMLFile.Docs, doc)
	if err := urlPath.ReplaceWithNode(&tmpYAMLFile, nodeToWrite); err != nil {
		return false, fmt.Errorf("replacing yaml key %q: %w", key, err)
	}
	yamlFile.Docs[index].Body = tmpYAMLFile.Docs[0].Body

	resultTarget.Description = fmt.Sprintf("%s\nkey %q%supdated from %q to %q, in file %q",
		resultTarget.Description,
		key,
		shouldMessage(dryRun),
		oldVersion,
		valueToWrite,
		originFilePath)

	return true, nil
}

// shouldMessage crafts the message fragment depending on whether Updatecli runs in
// dry run mode or not.
func shouldMessage(dryRun bool) string {
	if dryRun {
		return " should be "
	}
	return " "
}

package yaml

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/result"

	goyaml "github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/goccy/go-yaml/token"
)

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

package yaml

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"

	"github.com/updatecli/updatecli/pkg/core/result"

	"go.yaml.in/yaml/v3"
)

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

		// Decode the file into one or more YAML document nodes
		var docs []*yaml.Node
		dec := yaml.NewDecoder(strings.NewReader(y.files[filePath].content))
		for {
			var doc yaml.Node
			if derr := dec.Decode(&doc); derr != nil {
				if derr == io.EOF {
					break
				}
				return 0, ignoredFiles, fmt.Errorf("parsing yaml file %q: %w", originFilePath, derr)
			}
			docs = append(docs, &doc)
		}

		// If the file has no documents, treat as ignored
		if len(docs) == 0 {
			ignoredFiles++
			continue
		}

		// Process each key for this file
		for _, key := range keys {
			urlPath, err := yamlpath.NewPath(y.spec.Key)
			if err != nil {
				return 0, 0, fmt.Errorf("crafting yamlpath query for key %q: %w", key, err)
			}

			// No DocumentIndex: search across all documents and process each doc that matches
			foundAny := false
			var docNotChangedCount int

			for index, doc := range docs {

				if y.spec.DocumentIndex != nil {
					if index != *y.spec.DocumentIndex {
						continue
					}
				}

				nodes, err := urlPath.Find(doc)
				if err != nil {
					return 0, ignoredFiles, fmt.Errorf("searching for key %q in yaml file: %w", key, err)
				}
				if len(nodes) == 0 {
					continue
				}

				foundAny = true
				fileKeysProcessed++
				var oldVersion string
				var notChangedNode int
				for _, node := range nodes {
					oldVersion = node.Value
					resultTarget.Information = oldVersion

					if oldVersion == valueToWrite {
						resultTarget.Description = fmt.Sprintf("%s\nkey %q already set to %q, from file %q (doc %d)",
							resultTarget.Description,
							key,
							valueToWrite,
							originFilePath,
							index)
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
						shouldMsg = " should be "
					}

					resultTarget.Description = fmt.Sprintf("%s\nkey %q%supdated from %q to %q, in file %q (doc %d)",
						resultTarget.Description,
						key,
						shouldMsg,
						oldVersion,
						valueToWrite,
						originFilePath,
						index)
				}
				if notChangedNode == len(nodes) {
					docNotChangedCount++
				}

				if !foundAny {
					if y.spec.SearchPattern {
						logrus.Debugf("ignoring key %q in file %q as we couldn't find it in any document", key, originFilePath)
						continue
					}
					return 0, ignoredFiles, fmt.Errorf("couldn't find key %q from file %q", key, originFilePath)
				}
				if docNotChangedCount == len(docs) {
					// if every matching document had only unchanged nodes, consider file not changed for this key
					fileNotChanged++
				}
			}
		} // end keys loop

		// If no keys were processed for this file (all were ignored), count as ignored
		if fileKeysProcessed == 0 {
			ignoredFiles++
			continue
		}

		// If all processed keys in this file were unchanged, count as not changed
		if fileNotChanged == fileKeysProcessed {
			notChanged++
		}

		// Re-encode all documents back into buffer
		buf = bytes.Buffer{}
		for _, doc := range docs {
			if err := e.Encode(doc); err != nil {
				return 0, ignoredFiles, fmt.Errorf("unable to marshal the yaml file: %w", err)
			}
		}

		f := y.files[filePath]
		f.content = buf.String()
		// preserve leading document marker if it was present originally
		if strings.HasPrefix(y.files[filePath].content, "---\n") && !strings.HasPrefix(f.content, "---\n") {
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

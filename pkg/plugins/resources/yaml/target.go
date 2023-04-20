package yaml

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
	"gopkg.in/yaml.v3"
)

// Target updates a scm repository based on the modified yaml file.
func (y *Yaml) Target(source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {
	joignedFiles := make(map[string]string)
	for filePath := range y.files {
		joignedFilePath := filePath
		if scm != nil {
			joignedFilePath = joinPathWithWorkingDirectoryPath(joignedFilePath, scm.GetDirectory())
			logrus.Debugf("Relative path detected: changing from %q to absolute path from SCM: %q", filePath, joignedFilePath)
		}
		joignedFiles[joignedFilePath] = y.files[filePath]
	}
	y.files = joignedFiles

	return y.target(source, dryRun, resultTarget)
}

func (y *Yaml) target(source string, dryRun bool, resultTarget *result.Target) error {

	// Test if target reference a file with a prefix like https:// or file://, as we don't know how to update those files.
	for filePath := range y.files {
		if text.IsURL(filePath) {
			return fmt.Errorf("unsupported filename prefix for %s", filePath)
		}

		if strings.HasPrefix(filePath, "https://") ||
			strings.HasPrefix(filePath, "http://") {
			return fmt.Errorf("URL scheme is not supported for YAML target: %q", filePath)
		}

		// Test at runtime if a file exist (no ForceCreate for kind: yaml)
		if !y.contentRetriever.FileExists(filePath) {
			return fmt.Errorf("the yaml file %q does not exist", filePath)
		}
	}

	if err := y.Read(); err != nil {
		return err
	}

	valueToWrite := source
	if y.spec.Value != "" {
		valueToWrite = y.spec.Value
		logrus.Info("INFO: Using spec.Value instead of source input value.")
	}

	resultTarget.NewInformation = valueToWrite

	shouldMsg := "should be"

	// loop over file(s)
	notChanged := 0
	originalContents := make(map[string]string)
	for filePath := range y.files {
		originalContents[filePath] = y.files[filePath]

		out := yaml.Node{}
		err := yaml.Unmarshal([]byte(y.files[filePath]), &out)

		if err != nil {
			return fmt.Errorf("cannot unmarshal content of file %s: %v", filePath, err)
		}

		keyFound, oldVersion, _ := replace(&out, parseKey(y.spec.Key), valueToWrite, 1)

		resultTarget.OldInformation = oldVersion

		if !keyFound {
			return fmt.Errorf("couldn't find key %q from file %q",
				y.spec.Key,
				filePath)
		}

		if oldVersion == valueToWrite {
			resultTarget.Result = result.SUCCESS
			resultTarget.Description = fmt.Sprintf("%s\nkey %q already set to %q, from file %q, ",
				resultTarget.Description,
				y.spec.Key,
				valueToWrite,
				filePath)
			notChanged++
			continue
		}

		buf := new(bytes.Buffer)
		encoder := yaml.NewEncoder(buf)
		defer encoder.Close()
		encoder.SetIndent(y.indent)
		err = encoder.Encode(&out)

		if err != nil {
			return err
		}
		y.files[filePath] = buf.String()
		buf.Reset()

		resultTarget.Changed = true
		resultTarget.Files = append(resultTarget.Files, filePath)
		resultTarget.Result = result.ATTENTION

		logrus.Infof("%s\nkey %q %supdated from %q to %q, in file %q",
			resultTarget.Description,
			y.spec.Key,
			oldVersion,
			shouldMsg,
			valueToWrite,
			filePath)

		if !dryRun {
			newFile, err := os.Create(filePath)

			// https://staticcheck.io/docs/checks/#SA5001
			//lint:ignore SA5001 We want to defer the file closing before exiting the function
			defer newFile.Close()

			if err != nil {
				return err
			}

			err = y.contentRetriever.WriteToFile(y.files[filePath], filePath)
			if err != nil {
				return err
			}
		}
	}

	resultTarget.Description = strings.TrimPrefix(resultTarget.Description, "\n")

	// If no file was updated, don't return an error
	if notChanged == len(y.files) {
		return nil
	}

	sort.Strings(resultTarget.Files)

	return nil
}

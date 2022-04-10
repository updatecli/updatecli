package yaml

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
	"gopkg.in/yaml.v3"
)

var (
	yamlIndent int = 2
)

// Target updates a scm repository based on the modified yaml file.
func (y *Yaml) Target(source string, dryRun bool) (bool, error) {
	changed, _, _, err := y.target(source, dryRun)
	return changed, err
}

// TargetFromSCM updates a scm repository based on the modified yaml file.
func (y *Yaml) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (bool, []string, string, error) {
	absoluteFiles := make(map[string]string)
	for filePath := range y.files {
		absoluteFilePath := filePath
		if !filepath.IsAbs(filePath) {
			absoluteFilePath = filepath.Join(scm.GetDirectory(), filePath)
			logrus.Debugf("Relative path detected: changing to absolute path from SCM: %q", absoluteFilePath)
		}
		absoluteFiles[absoluteFilePath] = y.files[filePath]
	}
	y.files = absoluteFiles

	return y.target(source, dryRun)
}

func (y *Yaml) target(source string, dryRun bool) (bool, []string, string, error) {
	var files []string
	var message strings.Builder

	// Test if target reference a file with a prefix like https:// or file://, as we don't know how to update those files.
	for filePath := range y.files {
		if text.IsURL(filePath) {
			return false, files, message.String(), fmt.Errorf("unsupported filename prefix for %s", filePath)
		}
		// Test at runtime if a file exist (no ForceCreate for kind: yaml)
		if !y.contentRetriever.FileExists(filePath) {
			return false, files, message.String(), fmt.Errorf("the yaml file %q does not exist", filePath)
		}
	}

	if err := y.Read(); err != nil {
		return false, files, message.String(), err
	}

	valueToWrite := source
	if y.spec.Value != "" {
		valueToWrite = y.spec.Value
		logrus.Info("INFO: Using spec.Value instead of source input value.")
	}

	// loop over file(s)
	notChanged := 0
	originalContents := make(map[string]string)
	for filePath := range y.files {
		originalContents[filePath] = y.files[filePath]

		out := yaml.Node{}

		err := yaml.Unmarshal([]byte(y.files[filePath]), &out)

		if err != nil {
			return false, files, message.String(), fmt.Errorf("cannot unmarshal content of file %s: %v", filePath, err)
		}

		keyFound, oldVersion, _ := replace(&out, strings.Split(y.spec.Key, "."), valueToWrite, 1)

		if !keyFound {
			return false, files, message.String(), fmt.Errorf("%s cannot find key '%s' from file '%s'",
				result.FAILURE,
				y.spec.Key,
				filePath)
		}

		if oldVersion == valueToWrite {
			logrus.Infof("%s Key '%s', from file '%v', already set to %s, nothing else need to be done",
				result.SUCCESS,
				y.spec.Key,
				filePath,
				valueToWrite)
			notChanged++
		} else {
			newFileContent, err := yaml.Marshal(&out)
			if err != nil {
				return false, files, message.String(), err
			}
			y.files[filePath] = string(newFileContent)

			files = append(files, filePath)

			logrus.Infof("%s Key '%s', from file '%v', was updated from '%s' to '%s'",
				result.ATTENTION,
				y.spec.Key,
				filePath,
				oldVersion,
				valueToWrite)
		}

		if !dryRun {

			newFile, err := os.Create(filePath)

			// https://staticcheck.io/docs/checks/#SA5001
			//lint:ignore SA5001 We want to defer the file closing before exiting the function
			defer newFile.Close()

			if err != nil {
				return false, files, message.String(), nil
			}

			encoder := yaml.NewEncoder(newFile)
			defer encoder.Close()
			encoder.SetIndent(yamlIndent)
			err = encoder.Encode(&out)
			if err != nil {
				return false, files, message.String(), err
			}
			// TODO: check if all this encoder stuff is needed, and why yamlIndent isn't respected
			err = y.contentRetriever.WriteToFile(y.files[filePath], filePath)
			if err != nil {
				return false, files, message.String(), err
			}
		}
		message.WriteString(fmt.Sprintf("Update key %q from file %q", y.spec.Key, filePath))

	}

	// If no file was updated, return an error
	// TODO: why?
	if notChanged == len(y.files) {
		return false, files, message.String(), fmt.Errorf("no file was updated")
	}

	sort.Strings(files)

	return true, files, message.String(), nil
}

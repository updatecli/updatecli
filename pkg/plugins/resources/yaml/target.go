package yaml

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
	"gopkg.in/yaml.v3"
)

// Target updates a scm repository based on the modified yaml file.
func (y *Yaml) Target(source string, dryRun bool) (bool, error) {
	changed, _, _, err := y.target(source, dryRun)
	return changed, err
}

// TargetFromSCM updates a scm repository based on the modified yaml file.
func (y *Yaml) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (bool, []string, string, error) {
	if !filepath.IsAbs(y.spec.File) {
		y.spec.File = filepath.Join(scm.GetDirectory(), y.spec.File)
	}
	return y.target(source, dryRun)
}

func (y *Yaml) target(source string, dryRun bool) (bool, []string, string, error) {
	var files []string
	var message string

	// Test at runtime if a file exist
	if !y.contentRetriever.FileExists(y.spec.File) {
		return false, files, message, fmt.Errorf("the yaml file %q does not exist", y.spec.File)
	}

	if text.IsURL(y.spec.File) {
		return false, files, message, fmt.Errorf("unsupported filename prefix")
	}

	if err := y.Read(); err != nil {
		return false, files, message, err
	}
	data := y.currentContent

	out := yaml.Node{}

	err := yaml.Unmarshal([]byte(data), &out)

	if err != nil {
		return false, files, message, fmt.Errorf("cannot unmarshal data: %v", err)
	}

	valueToWrite := source
	if y.spec.Value != "" {
		valueToWrite = y.spec.Value
		logrus.Info("INFO: Using spec.Value instead of source input value.")
	}

	keyFound, oldVersion, _ := replace(&out, strings.Split(y.spec.Key, "."), valueToWrite, 1)

	if !keyFound {
		return false, files, message, fmt.Errorf("%s cannot find key '%s' from file '%s'",
			result.FAILURE,
			y.spec.Key,
			y.spec.File)
	}

	if oldVersion == valueToWrite {
		logrus.Infof("%s Key '%s', from file '%v', already set to %s, nothing else need to be done",
			result.SUCCESS,
			y.spec.Key,
			y.spec.File,
			valueToWrite)
		return false, files, message, nil
	}
	logrus.Infof("%s Key '%s', from file '%v', was updated from '%s' to '%s'",
		result.ATTENTION,
		y.spec.Key,
		y.spec.File,
		oldVersion,
		valueToWrite)

	if !dryRun {

		newFile, err := os.Create(y.spec.File)

		// https://staticcheck.io/docs/checks/#SA5001
		//lint:ignore SA5001 We want to defer the file closing before exiting the function
		defer newFile.Close()

		if err != nil {
			return false, files, message, nil
		}

		encoder := yaml.NewEncoder(newFile)
		defer encoder.Close()
		encoder.SetIndent(yamlIdent)
		err = encoder.Encode(&out)

		if err != nil {
			return false, files, message, err
		}
	}

	files = append(files, y.spec.File)

	message = fmt.Sprintf("Update key %q from file %q", y.spec.Key, y.spec.File)

	return true, files, message, nil
}

package yaml

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
	"gopkg.in/yaml.v3"
)

// Target updates a scm repository based on the modified yaml file.
func (y *Yaml) Target(source string, dryRun bool) (bool, error) {
	if strings.HasPrefix(y.spec.File, "https://") ||
		strings.HasPrefix(y.spec.File, "http://") {
		return false, fmt.Errorf("URL scheme is not supported for YAML target: %q", y.spec.File)
	}

	// Test at runtime if a file exist
	if !y.contentRetriever.FileExists(y.spec.File) {
		return false, fmt.Errorf("the yaml file %q does not exist", y.spec.File)
	}

	if text.IsURL(y.spec.File) {
		return false, fmt.Errorf("unsupported filename prefix")
	}

	if err := y.Read(); err != nil {
		return false, err
	}
	data := y.currentContent

	out := yaml.Node{}

	err := yaml.Unmarshal([]byte(data), &out)

	if err != nil {
		return false, fmt.Errorf("cannot unmarshal data: %v", err)
	}

	valueToWrite := source
	if y.spec.Value != "" {
		valueToWrite = y.spec.Value
		logrus.Info("INFO: Using spec.Value instead of source input value.")
	}

	keyFound, oldVersion, _ := replace(&out, parseKey(y.spec.Key), valueToWrite, 1)

	if !keyFound {
		return false, fmt.Errorf("%s cannot find key '%s' from file '%s'",
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
		return false, nil
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
			return false, nil
		}

		encoder := yaml.NewEncoder(newFile)
		defer encoder.Close()
		encoder.SetIndent(yamlIdent)
		err = encoder.Encode(&out)

		if err != nil {
			return false, err
		}
	}
	return true, nil
}

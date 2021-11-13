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
func (y *Yaml) TargetFromSCM(source string, scm scm.Scm, dryRun bool) (bool, []string, string, error) {
	if !filepath.IsAbs(y.Spec.File) {
		y.Spec.File = filepath.Join(scm.GetDirectory(), y.Spec.File)
	}
	return y.target(source, dryRun)
}

func (y *Yaml) target(source string, dryRun bool) (bool, []string, string, error) {
	var files []string
	var message string

	if text.IsURL(y.Spec.File) {
		return false, files, message, fmt.Errorf("unsupported filename prefix")
	}

	if err := y.Read(); err != nil {
		return false, files, message, err
	}
	data := y.CurrentContent

	out := yaml.Node{}

	err := yaml.Unmarshal([]byte(data), &out)

	if err != nil {
		return false, files, message, fmt.Errorf("cannot unmarshal data: %v", err)
	}

	valueToWrite := source
	if y.Spec.Value != "" {
		valueToWrite = y.Spec.Value
		logrus.Info("INFO: Using spec.Value instead of source input value.")
	}

	keyFound, oldVersion, _ := replace(&out, strings.Split(y.Spec.Key, "."), valueToWrite, 1)

	if !keyFound {
		return false, files, message, fmt.Errorf("%s cannot find key '%s' from file '%s'",
			result.FAILURE,
			y.Spec.Key,
			y.Spec.File)
	}

	if oldVersion == valueToWrite {
		logrus.Infof("%s Key '%s', from file '%v', already set to %s, nothing else need to be done",
			result.SUCCESS,
			y.Spec.Key,
			y.Spec.File,
			valueToWrite)
		return false, files, message, nil
	}
	logrus.Infof("%s Key '%s', from file '%v', was updated from '%s' to '%s'",
		result.ATTENTION,
		y.Spec.Key,
		y.Spec.File,
		oldVersion,
		valueToWrite)

	if !dryRun {

		newFile, err := os.Create(y.Spec.File)
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

	files = append(files, y.Spec.File)

	message = fmt.Sprintf("Update key %q from file %q", y.Spec.Key, y.Spec.File)

	return true, files, message, nil
}

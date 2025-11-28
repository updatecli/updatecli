package yaml

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/text"
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

package lock

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (t *TerraformLock) Target(source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {
	if scm != nil {
		t.UpdateAbsoluteFilePath(scm.GetDirectory())
	}

	for _, resourceFile := range t.files {
		// Target doesn't support updating files on remote http location
		if strings.HasPrefix(resourceFile.filePath, "https://") ||
			strings.HasPrefix(resourceFile.filePath, "http://") {
			return fmt.Errorf("%s URL scheme is not supported for HCL target: %q", result.FAILURE, t.spec.File)
		}
	}

	if err := t.Read(); err != nil {
		return err
	}

	address := t.spec.Provider

	valueToWrite := source
	if t.spec.Value != "" {
		valueToWrite = t.spec.Value
		logrus.Debug("Using spec.Value instead of source input value.")
	}

	resultTarget.NewInformation = valueToWrite

	notChanged := 0

	remoteHashes, err := t.getProviderHashes(valueToWrite)
	if err != nil {
		return err
	}

	for fileKey, resourceFile := range t.files {

		currentValue, currentHashes, err := t.Query(resourceFile)
		if err != nil {
			return err
		}

		resultTarget.Information = currentValue

		if currentValue == valueToWrite && reflect.DeepEqual(currentHashes, remoteHashes) {
			resultTarget.Description = fmt.Sprintf("%q already set to %q, from file %q, ",
				address,
				valueToWrite,
				resourceFile.originalFilePath)
			notChanged++
			continue
		}

		resultTarget.Changed = true
		resultTarget.Files = append(resultTarget.Files, resourceFile.originalFilePath)
		resultTarget.Result = result.ATTENTION
		resultTarget.Description = fmt.Sprintf("%q updated from %q to %q in file %q",
			address,
			currentValue,
			valueToWrite,
			resourceFile.originalFilePath)

		if !dryRun {

			if err := t.Apply(fileKey, valueToWrite, remoteHashes); err != nil {
				return err
			}

			if err := t.contentRetriever.WriteToFile(
				t.files[fileKey].content,
				t.files[fileKey].filePath,
			); err != nil {
				return err
			}

		}
	}

	if notChanged == len(t.files) {
		resultTarget.Result = result.SUCCESS
		return nil
	}

	sort.Strings(resultTarget.Files)

	return nil
}

package hcl

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (h *Hcl) Target(source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {
	if scm != nil {
		h.UpdateAbsoluteFilePath(scm.GetDirectory())
	}

	for _, f := range h.files {
		resourceFile := f

		// Target doesn't support updating files on remote http location
		if strings.HasPrefix(resourceFile.filePath, "https://") ||
			strings.HasPrefix(resourceFile.filePath, "http://") {
			return fmt.Errorf("%s URL scheme is not supported for HCL target: %q", result.FAILURE, h.spec.File)
		}
	}

	if err := h.Read(); err != nil {
		return fmt.Errorf("reading hcl file: %w", err)
	}

	query := h.spec.Path

	valueToWrite := source
	if h.spec.Value != "" {
		valueToWrite = h.spec.Value
		logrus.Debug("Using spec.Value instead of source input value.")
	}

	resultTarget.NewInformation = valueToWrite

	notChanged := 0

	var descriptions []string

	for fileKey, f := range h.files {
		resourceFile := f

		currentValue, err := h.Query(resourceFile)
		if err != nil {
			return err
		}

		resultTarget.Information = currentValue

		if currentValue == valueToWrite {
			descriptions = append(descriptions,
				fmt.Sprintf("path %q already set to %q, from file %q, ",
					query,
					valueToWrite,
					resourceFile.originalFilePath))
			notChanged++
			continue
		}

		resultTarget.Files = append(resultTarget.Files, resourceFile.originalFilePath)
		descriptions = append(descriptions,
			fmt.Sprintf("path %q updated from %q to %q in file %q",
				query,
				currentValue,
				valueToWrite,
				resourceFile.originalFilePath))

		if !dryRun {

			if err := h.Apply(fileKey, valueToWrite); err != nil {
				return err
			}

			if err := h.contentRetriever.WriteToFile(
				h.files[fileKey].content,
				h.files[fileKey].filePath,
			); err != nil {
				return err
			}

		}
	}

	sort.Strings(descriptions)
	descriptionLines := strings.Join(descriptions, "\n\t")

	if notChanged == len(h.files) {
		resultTarget.Result = result.SUCCESS
		resultTarget.Description = fmt.Sprintf("no changes detected:\n\t%s", descriptionLines)
		return nil
	}

	resultTarget.Changed = true
	resultTarget.Result = result.ATTENTION
	resultTarget.Description = fmt.Sprintf("changes detected:\n\t%s", descriptionLines)

	sort.Strings(resultTarget.Files)

	return nil
}

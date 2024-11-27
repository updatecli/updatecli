package dockerfile

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Target updates a targeted Dockerfile from source control management system
func (d *Dockerfile) Target(source result.SourceInformation, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) (err error) {
	// At the moment, this plugin do not return the currently used value
	// This could be a useful improvement for the source
	resultTarget.Information = "unknown"
	resultTarget.NewInformation = source.Value
	resultTarget.Changed = false
	resultTarget.Result = result.SUCCESS

	changeDescriptions := []string{}

	for _, file := range d.files {
		if !filepath.IsAbs(file) && scm != nil {
			file = filepath.Join(scm.GetDirectory(), file)
			logrus.Debugf("Relative path detected: changing to absolute path from SCM: %q", file)
		}

		dockerfileContent, err := d.contentRetriever.ReadAll(file)
		if err != nil {
			return err
		}

		logrus.Debugf("\n🐋 On (Docker)file %q:\n\n", file)

		newDockerfileContent, changedLines, err := d.parser.ReplaceInstructions([]byte(dockerfileContent), source.Value, d.spec.Stage)
		if err != nil {
			return err
		}

		if len(changedLines) == 0 {
			logrus.Debugf("no change detected %q, nothing else to do", file)
		} else {
			resultTarget.Changed = true
		}

		lines := []int{}
		for idx := range changedLines {
			lines = append(lines, idx)
		}
		sort.Ints(lines)

		changeDescriptions = append(changeDescriptions, fmt.Sprintf("changed lines %v of file %q", lines, file))
		resultTarget.Files = append(resultTarget.Files, file)

		if !dryRun {
			// Write the new Dockerfile content from buffer to file
			err := d.contentRetriever.WriteToFile(string(newDockerfileContent), file)
			if err != nil {
				return err
			}
		}
	}

	if resultTarget.Changed {
		resultTarget.Result = result.ATTENTION
	}

	if len(changeDescriptions) > 0 {
		resultTarget.Description = strings.Join(changeDescriptions, ", ")
	}
	return err
}

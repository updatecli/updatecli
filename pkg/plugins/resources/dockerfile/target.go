package dockerfile

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Target updates a targeted Dockerfile from source control management system
func (d *Dockerfile) Target(source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) (err error) {
	if !filepath.IsAbs(d.spec.File) && scm != nil {
		d.spec.File = filepath.Join(scm.GetDirectory(), d.spec.File)
		logrus.Debugf("Relative path detected: changing to absolute path from SCM: %q", d.spec.File)
	}
	return d.target(source, dryRun, resultTarget)
}

func (d *Dockerfile) target(source string, dryRun bool, resultTarget *result.Target) (err error) {
	dockerfileContent, err := d.contentRetriever.ReadAll(d.spec.File)
	if err != nil {
		return err
	}

	logrus.Debugf("\n🐋 On (Docker)file %q:\n\n", d.spec.File)

	// At the moment, this plugin do not return the currently used value
	// This could be a useful improvement for the source
	resultTarget.OldInformation = "unknown"
	resultTarget.NewInformation = source

	newDockerfileContent, changedLines, err := d.parser.ReplaceInstructions([]byte(dockerfileContent), source)
	if err != nil {
		return err
	}

	if len(changedLines) == 0 {
		logrus.Debugf("empty file detected %q, nothing to do", d.spec.File)
		return nil
	}

	lines := []int{}
	for idx := range changedLines {
		lines = append(lines, idx)
	}
	sort.Ints(lines)

	resultTarget.Description = fmt.Sprintf("changed lines %v of file %q", lines, d.spec.File)
	resultTarget.Files = append(resultTarget.Files, d.spec.File)
	resultTarget.Changed = true
	resultTarget.Result = result.ATTENTION

	if !dryRun {
		// Write the new Dockerfile content from buffer to file
		err := d.contentRetriever.WriteToFile(string(newDockerfileContent), d.spec.File)
		if err != nil {
			return err
		}
	}

	return err
}

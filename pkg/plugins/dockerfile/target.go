package dockerfile

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/helpers"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Target updates a targeted Dockerfile
func (d *Dockerfile) Target(source string, dryRun bool) (bool, error) {
	changed, _, _, err := d.target(source, dryRun)
	return changed, err
}

// TargetFromSCM updates a targeted Dockerfile from source controle management system
func (d *Dockerfile) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (changed bool, files []string, message string, err error) {
	if !filepath.IsAbs(d.spec.File) {
		d.spec.File = filepath.Join(scm.GetDirectory(), d.spec.File)
		logrus.Debugf("Relative path detected: changing to absolute path from SCM: %q", d.spec.File)
	}
	return d.target(source, dryRun)
}

func (d *Dockerfile) target(source string, dryRun bool) (changed bool, files []string, message string, err error) {
	dockerfileContent, err := helpers.ReadFile(d.spec.File)
	if err != nil {
		return false, files, message, err
	}

	logrus.Infof("\n🐋 On (Docker)file %q:\n\n", d.spec.File)

	newDockerfileContent, changedLines, err := d.parser.ReplaceInstructions(dockerfileContent, source)
	if err != nil {
		return false, files, message, err
	}

	if len(changedLines) == 0 {
		return false, files, message, err
	}

	lines := []int{}
	for idx := range changedLines {
		lines = append(lines, idx)
	}

	message = fmt.Sprintf("changed lines %v of file %q", lines, d.spec.File)
	files = append(files, d.spec.File)

	if !dryRun {
		// Write the new Dockerfile content from buffer to file
		err = ioutil.WriteFile(d.spec.File, newDockerfileContent, 0600)
		if err != nil {
			log.Fatal(err)
		}
	}

	return true, files, message, err
}

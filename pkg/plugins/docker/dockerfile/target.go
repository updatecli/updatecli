package dockerfile

import (
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/helpers"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Target updates a targeted Dockerfile
func (d *Dockerfile) Target(source string, dryRun bool) (bool, error) {

	err := d.SetParser()
	if err != nil {
		return false, err
	}

	// read Dockerfile content
	dockerfileContent, err := helpers.ReadFile(d.File)
	if err != nil {
		return false, err
	}

	logrus.Infof("\n🐋 On (Docker)file %q:\n\n", d.File)

	newDockerfileContent, changedLines, err := d.parser.ReplaceInstructions(dockerfileContent, source)
	if err != nil {
		return false, err
	}

	if len(changedLines) == 0 {
		return false, nil
	}

	if !dryRun {
		// Write the new Dockerfile content from buffer to file
		err = ioutil.WriteFile(d.File, newDockerfileContent, 0600)
		if err != nil {
			log.Fatal(err)
		}
	}

	return true, nil
}

// TargetFromSCM updates a targeted Dockerfile from source controle management system
func (d *Dockerfile) TargetFromSCM(source string, scm scm.Scm, dryRun bool) (changed bool, files []string, message string, err error) {

	file := d.File
	d.File = path.Join(scm.GetDirectory(), d.File)

	changed, err = d.Target(source, d.DryRun)
	if err != nil {
		return changed, files, message, err

	}

	files = append(files, file)

	if changed {
		// Generate a multiline string with one message per line (each message maps to a line changed by the parser)
		messageList := strings.Join(d.messages, "\n")
		// Generate a nice commit message with a first line title, and append the previous message list
		message = fmt.Sprintf("Update Dockerfile instruction for %q\n%s\n", d.File, messageList)
	}

	return changed, files, message, err
}

package dockerfile

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/olblak/updateCli/pkg/core/helpers"
	"github.com/olblak/updateCli/pkg/core/scm"
)

// Target updates a targeted Dockerfile
func (d *Dockerfile) Target(source string, dryRun bool) (changed bool, err error) {

	d.Value = source

	changed = false

	raw, err := helpers.ReadFile(d.File)

	if err != nil {
		return changed, err
	}

	data, err := parser.Parse(bytes.NewReader(raw))

	if err != nil {
		return false, err
	}

	valueFound, oldVersion, err := d.replace(data.AST)

	if err != nil {
		return changed, err
	}

	if valueFound {
		if oldVersion == d.Value {
			logrus.Infof("\u2714 Instruction '%s', from Dockerfile '%v', already set to %s, nothing else need to be done",
				d.Instruction,
				d.File,
				d.Value)
			return changed, nil
		}

		changed = true
		logrus.Infof("\u2714 Instruction '%s', from Dockerfile '%v', was updated from '%s' to '%s'",
			d.Instruction,
			d.File,
			oldVersion,
			d.Value)

	} else {
		logrus.Infof("\u2717 cannot find instruction '%s' from Dockerfile '%s'", d.Instruction, d.File)
		return changed, nil
	}

	if !dryRun {

		newFile, err := os.Create(d.File)
		defer newFile.Close()

		if err != nil {
			return changed, fmt.Errorf("something went wrong while encoding %v", err)
		}

		document := ""

		err = Marshal(data, &document)
		if err != nil {
			return changed, err
		}

		writer := bufio.NewWriter(newFile)

		for _, line := range strings.Split(document, "\n") {
			_, err := writer.WriteString(line + "\n")
			if err != nil {
				log.Fatalf("Got error while writing to a file. Err: %s", err.Error())
			}
		}

		writer.Flush()

		err = newFile.Close()
		if err != nil {
			return changed, err
		}

	}

	return changed, nil
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
		message = fmt.Sprintf("[updatecli] Instruction '%s' from Dockerfile '%s' updated to '%s'\n",
			d.Instruction,
			file,
			d.Value,
		)
	}

	return changed, files, message, err
}

package dockerfile

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"path"

	"github.com/olblak/updateCli/pkg/core/helpers"
	"github.com/olblak/updateCli/pkg/core/scm"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

// Condition test if the Dockerfile contains the correct key/value
func (d *Dockerfile) Condition(version string) (bool, error) {

	raw, err := helpers.ReadFile(d.File)
	if err != nil {
		return false, err
	}

	if len(d.Value) == 0 {
		d.Value = version
	}

	data, err := parser.Parse(bytes.NewReader(raw))

	if err != nil {
		return false, err
	}

	found, val, err := d.replace(data.AST)

	if err != nil {
		return false, err
	}

	if found {
		if val == d.Value {
			logrus.Infof("\u2714 Instruction '%s' from Dockerfile '%s', is correctly set to '%s' \n",
				d.Instruction,
				d.File,
				d.Value)
			return true, nil
		}

		logrus.Infof("\u2717 Instruction '%s' from Dockerfile '%s', is incorrectly set to '%s' instead of '%s'\n",
			d.Instruction,
			d.File,
			val,
			d.Value)

	} else {

		logrus.Infof("\u2717 Instruction '%s' from Dockerfile '%s', wasn't found \n", d.Instruction, d.File)
	}

	return false, nil

}

// ConditionFromSCM run based on a file from SCM
func (d *Dockerfile) ConditionFromSCM(version string, scm scm.Scm) (bool, error) {

	raw, err := helpers.ReadFile(path.Join(scm.GetDirectory(), d.File))

	if err != nil {
		return false, err
	}

	if len(d.Value) == 0 {
		d.Value = version
	}

	data, err := parser.Parse(bytes.NewReader(raw))

	if err != nil {
		return false, err
	}

	found, val, err := d.replace(data.AST)

	if err != nil {
		return false, err
	}

	if found {
		if val == d.Value {
			logrus.Infof("\u2714 Instruction '%s' from Dockerfile '%s', is correctly set to '%s' \n",
				d.Instruction,
				d.File,
				d.Value)
			return true, nil
		}

		logrus.Infof("\u2717 Instruction '%s' from Dockerfile '%s', is incorrectly set to '%s' instead of '%s'\n",
			d.Instruction,
			d.File,
			val,
			d.Value)

	} else {
		logrus.Infof("\u2717 Instruction '%s' from Dockerfile '%s', wasn't found \n", d.Instruction, d.File)

	}

	return false, nil
}

package dockerfile

import (
	"path"

	"github.com/olblak/updateCli/pkg/core/helpers"
	"github.com/olblak/updateCli/pkg/core/scm"
	"github.com/sirupsen/logrus"
)

// Condition test if the Dockerfile contains the correct key/value
func (d *Dockerfile) Condition(source string) (bool, error) {

	err := d.SetParser()
	if err != nil {
		return false, err
	}
	// read Dockerfile content
	dockerfileContent, err := helpers.ReadFile(d.File)
	if err != nil {
		return false, err
	}

	found := d.parser.FindInstruction(dockerfileContent)

	if !found {
		logrus.Infof("\u2717 Instruction %v from Dockerfile %q, wasn't found", d.Instruction, d.File)
		return false, nil
	}

	logrus.Infof("\u2714 There is a match for the instruction %s in the Dockerfile %q", d.Instruction, d.File)

	return true, nil
}

// ConditionFromSCM run based on a file from SCM
func (d *Dockerfile) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {
	d.File = path.Join(scm.GetDirectory(), d.File)

	found, err := d.Condition(source)
	if err != nil {
		return false, err
	}

	return found, nil
}

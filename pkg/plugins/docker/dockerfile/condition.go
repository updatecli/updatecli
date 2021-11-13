package dockerfile

import (
	"path"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/helpers"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
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

	logrus.Infof("\n🐋 On (Docker)file %q:\n\n", d.File)

	found := d.parser.FindInstruction(dockerfileContent)

	return found, nil
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

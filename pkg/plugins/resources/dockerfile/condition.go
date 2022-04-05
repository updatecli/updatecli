package dockerfile

import (
	"fmt"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Condition test if the Dockerfile contains the correct key/value
func (d *Dockerfile) Condition(source string) (bool, error) {

	if !d.contentRetriever.FileExists(d.spec.File) {
		return false, fmt.Errorf("the file %s does not exist", d.spec.File)
	}
	dockerfileContent, err := d.contentRetriever.ReadAll(d.spec.File)
	if err != nil {
		return false, err
	}

	logrus.Infof("\n🐋 On (Docker)file %q:\n\n", d.spec.File)

	found := d.parser.FindInstruction([]byte(dockerfileContent))

	return found, nil
}

// ConditionFromSCM run based on a file from SCM
func (d *Dockerfile) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	d.spec.File = path.Join(scm.GetDirectory(), d.spec.File)

	found, err := d.Condition(source)
	if err != nil {
		return false, err
	}

	return found, nil
}

package dockerfile

import (
	"fmt"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition test if the Dockerfile contains the correct key/value
func (d *Dockerfile) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {

	if scm != nil {
		d.spec.File = path.Join(scm.GetDirectory(), d.spec.File)
	}

	if !d.contentRetriever.FileExists(d.spec.File) {
		return fmt.Errorf("the file %s does not exist", d.spec.File)
	}
	dockerfileContent, err := d.contentRetriever.ReadAll(d.spec.File)
	if err != nil {
		return fmt.Errorf("reading dockerfile: %w", err)
	}

	logrus.Debugf("\n🐋 On (Docker)file %q:\n\n", d.spec.File)

	found := d.parser.FindInstruction([]byte(dockerfileContent))

	switch found {
	case true:
		resultCondition.Pass = true
		resultCondition.Result = result.SUCCESS
		resultCondition.Description = fmt.Sprintf("key %q found in Dockerfile %q",
			d.spec.Instruction,
			d.spec.File,
		)
	case false:
		resultCondition.Pass = false
		resultCondition.Result = result.FAILURE
		resultCondition.Description = fmt.Sprintf("key %q not found in Dockerfile %q",
			d.spec.Instruction,
			d.spec.File,
		)
	}

	return nil
}

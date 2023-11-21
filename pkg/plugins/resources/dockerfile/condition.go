package dockerfile

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Condition test if the Dockerfile contains the correct key/value
func (d *Dockerfile) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	globalPass := true
	descriptionList := []string{}

	for _, file := range d.files {
		if !filepath.IsAbs(file) && scm != nil {
			file = path.Join(scm.GetDirectory(), file)
		}

		if !d.contentRetriever.FileExists(file) {
			return false, "", fmt.Errorf("the file %s does not exist", file)
		}
		dockerfileContent, err := d.contentRetriever.ReadAll(file)
		if err != nil {
			return false, "", fmt.Errorf("reading dockerfile: %w", err)
		}

		logrus.Debugf("\n🐋 On (Docker)file %q:\n\n", file)

		found := d.parser.FindInstruction([]byte(dockerfileContent))

		switch found {
		case true:
			globalPass = true && globalPass
			descriptionList = append(descriptionList, fmt.Sprintf("key %q found in Dockerfile %q",
				d.spec.Instruction,
				file,
			))
		case false:
			globalPass = false && globalPass
			descriptionList = append(descriptionList, fmt.Sprintf("key %q not found in Dockerfile %q",
				d.spec.Instruction,
				file,
			))
		}
	}

	return globalPass, strings.Join(descriptionList, ", "), nil
}

package dockerfile

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
)

// Condition test if the Dockerfile contains the correct key/value
func (d *Dockerfile) Condition(_ context.Context, source string, scm scm.ScmHandler, resolver utils.Resolver) (pass bool, message string, err error) {
	globalPass := true
	descriptionList := []string{}

	for _, relativeFile := range d.files {
		file, err := resolver.Resolve(relativeFile)
		if err != nil {
			return false, "", fmt.Errorf("invalid file path %q: %w", relativeFile, err)
		}

		if !d.contentRetriever.FileExists(file) {
			return false, "", fmt.Errorf("the file %s does not exist", file)
		}
		dockerfileContent, err := d.contentRetriever.ReadAll(file)
		if err != nil {
			return false, "", fmt.Errorf("reading dockerfile: %w", err)
		}

		logrus.Debugf("\n🐋 On (Docker)file %q:\n\n", file)

		found := d.parser.FindInstruction([]byte(dockerfileContent), d.spec.Stage)

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

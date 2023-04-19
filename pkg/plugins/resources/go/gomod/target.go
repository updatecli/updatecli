package gomod

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
)

// Target is not supported for the Golang resource
func (g *GoMod) Target(source string, scm scm.ScmHandler, dryRun bool) (changed bool, files []string, message string, err error) {

	version := source
	if g.spec.Version != "" {
		version = g.spec.Version
	}

	filename := g.filename
	if scm != nil {
		filename = utils.JoinFilePathWithWorkingDirectoryPath(g.filename, scm.GetDirectory())
	}

	oldVersion, newVersion, changed, err := g.setVersion(version, filename, dryRun)
	if err != nil {
		return false, files, message, err
	}

	if !changed {
		switch g.kind {
		case kindGolang:
			message = fmt.Sprintf("go.mod already set Golang version to %q", newVersion)
		case kindModule:
			message = fmt.Sprintf("go.mod already has Module %q set to version %q", g.spec.Module, newVersion)
		}

		logrus.Infoln(message)
		return changed, files, message, nil
	}

	if dryRun {
		switch g.kind {
		case kindGolang:
			message = fmt.Sprintf("go.mod should update Golang version from %q to %q", oldVersion, newVersion)
		case kindModule:
			message = fmt.Sprintf("go.mod should update Module path %q version from %q to %q", g.spec.Module, oldVersion, newVersion)
		}

		logrus.Infoln(message)
		return changed, files, message, nil
	}

	files = []string{filename}
	switch g.kind {
	case kindGolang:
		message = fmt.Sprintf("go.mod updated Golang version from %q to %q", oldVersion, newVersion)
	case kindModule:
		message = fmt.Sprintf("go.mod updated Module path %q version from %q to %q", g.spec.Module, oldVersion, newVersion)
	}

	logrus.Infoln(message)
	return changed, files, message, nil
}

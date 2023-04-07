package gomod

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"

	"github.com/updatecli/updatecli/pkg/plugins/utils"
)

// Source returns the latest go module version
func (g *GoMod) Source(workingDir string) (string, error) {
	var err error

	g.foundVersion, err = g.version(utils.JoinFilePathWithWorkingDirectoryPath(g.filename, workingDir))
	if err != nil {
		return "", err
	}

	if g.foundVersion == "" {
		err = fmt.Errorf("%s no version found for module path %q", result.FAILURE, g.spec.Module)
		return "", err
	}

	switch g.kind {
	case kindGolang:
		logrus.Infof("%s Golang Version %s found", result.SUCCESS, g.foundVersion)
	case kindModule:
		logrus.Infof("%s Version %s found for GO module %q", result.SUCCESS, g.foundVersion, g.spec.Module)
	}

	return g.foundVersion, nil
}

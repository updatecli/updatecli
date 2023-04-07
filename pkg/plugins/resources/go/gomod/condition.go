package gomod

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition checks that a specific golang version exists
func (g *GoMod) Condition(source string) (bool, error) {
	return g.condition(source)
}

// ConditionFromSCM checks that a specific golang version exists. SCM doesn't affect the result
func (g *GoMod) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	return g.condition(source)
}

// Condition checks if a specific stable Golang version is published
func (g *GoMod) condition(source string) (bool, error) {
	var err error

	versionToCheck := g.spec.Version
	if versionToCheck == "" {
		versionToCheck = source
	}

	if len(versionToCheck) == 0 {
		return false, errors.New("no version defined")
	}

	g.foundVersion, err = g.version(g.filename)
	if err != nil {

		if err == ErrModuleNotFound {
			logrus.Infof("%s module path %q not found\n",
				result.FAILURE, g.spec.Module)
			return false, nil
		}

		return false, err
	}

	if g.foundVersion == versionToCheck {
		switch g.kind {
		case kindGolang:
			logrus.Infof("%s Golang version %q found",
				result.SUCCESS, g.foundVersion)
		case kindModule:
			logrus.Infof("%s version %q found for module %q",
				result.SUCCESS, g.foundVersion, g.spec.Module)
		}
		return true, nil
	}

	switch g.kind {
	case kindGolang:
		logrus.Infof("%s Golang version %q found, expected %q",
			result.SUCCESS, g.foundVersion, versionToCheck)
	case kindModule:
		logrus.Infof("%s version %q found, expected %s for module path %q\n",
			result.FAILURE, g.foundVersion, versionToCheck, g.spec.Module)
	}

	return false, nil
}

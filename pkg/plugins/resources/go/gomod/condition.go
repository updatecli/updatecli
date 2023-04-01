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
	versionToCheck := g.spec.Version
	if versionToCheck == "" {
		versionToCheck = source
	}

	if len(versionToCheck) == 0 {
		return false, errors.New("no version defined")
	}

	version, err := g.version(g.filename)
	if err != nil {

		if err == ErrModuleNotFound {
			logrus.Infof("%s module path %q not found\n",
				result.FAILURE, g.spec.Module)
			return false, nil
		}

		return false, err
	}

	if version == versionToCheck {
		logrus.Infof("%s version %q found for module %q",
			result.SUCCESS, version, g.spec.Module)
		return true, nil
	}

	logrus.Infof("%s version %q found, expected %s for module path %q\n",
		result.FAILURE, version, versionToCheck, g.spec.Module)

	return false, nil
}

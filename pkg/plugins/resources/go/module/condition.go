package gomodule

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition checks that a version exists for a specific golang module
func (g *GoModule) Condition(source string) (bool, error) {
	return g.condition(source)
}

// ConditionFromSCM is not support supported
func (g *GoModule) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	return g.condition(source)
}

// Condition checks if a go module with a specific version is published
func (g *GoModule) condition(source string) (bool, error) {
	versionToCheck := g.Spec.Version
	if versionToCheck == "" {
		versionToCheck = source
	}
	if len(versionToCheck) == 0 {
		return false, errors.New("no version defined")
	}

	_, versions, err := g.versions()
	if err != nil {
		return false, err
	}

	for _, v := range versions {
		if v == versionToCheck {
			logrus.Infof("%s version %q available\n", result.SUCCESS, versionToCheck)
			return true, nil
		}
	}

	logrus.Infof("%s version %q doesn't exist\n", result.FAILURE, versionToCheck)

	return false, nil
}

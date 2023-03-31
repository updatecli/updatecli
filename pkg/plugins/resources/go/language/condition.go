package language

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition checks that a specific golang version exists
func (l *Language) Condition(source string) (bool, error) {
	return l.condition(source)
}

// ConditionFromSCM checks that a specific golang version exists. SCM doesn't affect the result
func (l *Language) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	return l.condition(source)
}

// Condition checks if a specific stable Golang version is published
func (l *Language) condition(source string) (bool, error) {
	versionToCheck := l.spec.Version
	if versionToCheck == "" {
		versionToCheck = source
	}
	if len(versionToCheck) == 0 {
		return false, errors.New("no version defined")
	}

	_, versions, err := l.versions()
	if err != nil {
		return false, err
	}

	for _, v := range versions {
		if v == versionToCheck {
			logrus.Infof("%s Golang version %q available\n", result.SUCCESS, versionToCheck)
			return true, nil
		}
	}

	logrus.Infof("%s Golang version %q doesn't exist\n", result.FAILURE, versionToCheck)

	return false, nil
}

package language

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Condition checks if a specific stable Golang version is published
func (l *Language) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	if scm != nil {
		logrus.Debugln("scm is not supported")
	}
	versionToCheck := l.Spec.Version
	if versionToCheck == "" {
		versionToCheck = source
	}
	if len(versionToCheck) == 0 {
		return false, "", fmt.Errorf("no version defined")
	}

	versions, err := l.versions()
	if err != nil {
		return false, "", fmt.Errorf("searching golang version: %w", err)
	}

	for _, v := range versions {
		if v == versionToCheck {
			return true, fmt.Sprintf("Golang version %q available", versionToCheck), nil
		}
	}

	return false, fmt.Sprintf("golang version %q doesn't exist", versionToCheck), nil
}

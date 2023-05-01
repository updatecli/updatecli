package language

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition checks if a specific stable Golang version is published
func (l *Language) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {
	if scm != nil {
		logrus.Debugln("scm is not supported")
	}
	versionToCheck := l.Spec.Version
	if versionToCheck == "" {
		versionToCheck = source
	}
	if len(versionToCheck) == 0 {
		return errors.New("no version defined")
	}

	versions, err := l.versions()
	if err != nil {
		return fmt.Errorf("searchin golang version: %w", err)
	}

	for _, v := range versions {
		if v == versionToCheck {
			resultCondition.Pass = true
			resultCondition.Result = result.SUCCESS
			resultCondition.Description = fmt.Sprintf("Golang version %q available", versionToCheck)
			return nil
		}
	}

	return fmt.Errorf("golang version %q doesn't exist", versionToCheck)
}

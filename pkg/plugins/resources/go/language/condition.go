package language

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Condition checks if a specific stable Golang version is published
func (l *Language) Condition(ctx context.Context, source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	var versions []string

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

	switch l.Spec.Age.IsZero() {
	case true:
		versions, err = l.versions(ctx)
		if err != nil {
			return false, "", fmt.Errorf("searching golang version: %w", err)
		}
	case false:
		versions, err = l.getTagsFromRepository()
		if err != nil {
			return false, "", fmt.Errorf("searching golang version: %w", err)
		}
	}

	for _, v := range versions {
		if v == versionToCheck {
			return true, fmt.Sprintf("Golang version %q available", versionToCheck), nil
		}
	}

	return false, fmt.Sprintf("golang version %q doesn't exist", versionToCheck), nil
}

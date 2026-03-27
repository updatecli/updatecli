package pypi

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Condition checks that a PyPI package version exists.
func (p *Pypi) Condition(source string, scmHandler scm.ScmHandler) (pass bool, message string, err error) {
	if scmHandler != nil {
		logrus.Warningf("SCM configuration is not supported for pypi condition, ignoring")
	}

	versionToCheck := p.spec.Version
	if versionToCheck == "" {
		versionToCheck = source
	}
	if versionToCheck == "" {
		return false, "", errors.New("no version defined")
	}

	_, versions, err := p.getVersions()
	if err != nil {
		return false, "", err
	}

	for _, v := range versions {
		if v == versionToCheck {
			return true, fmt.Sprintf("version %q exists for PyPI package %q", versionToCheck, p.spec.Name), nil
		}
	}

	return false, fmt.Sprintf("version %q does not exist for PyPI package %q", versionToCheck, p.spec.Name), nil
}

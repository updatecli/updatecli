package pypi

import (
	"context"
	"errors"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

// Condition checks that a PyPI package version exists.
func (p *Pypi) Condition(ctx context.Context, source string, scmHandler scm.ScmHandler, pathResolver pathresolver.Resolver) (pass bool, message string, err error) {
	if scmHandler != nil {
		log.Warningf("SCM configuration is not supported for pypi condition, ignoring")
	}

	versionToCheck := p.spec.Version
	if versionToCheck == "" {
		versionToCheck = source
	}
	if versionToCheck == "" {
		return false, "", errors.New("no version defined")
	}

	_, versions, err := p.getVersions(ctx)
	/*
		A cooldown still running means no published version is usable yet, which the loop
		below already reports as an unmet condition, so it isn't surfaced as an error.
	*/
	if err != nil && !errors.Is(err, age.ErrNoVersionMatchingAge) {
		return false, "", err
	}

	for _, v := range versions {
		if v == versionToCheck || p.originalVersion(v) == versionToCheck {
			return true, fmt.Sprintf("version %q exists for PyPI package %q", versionToCheck, p.spec.Name), nil
		}
	}

	if !p.spec.Age.IsZero() {
		return false, fmt.Sprintf("version %q does not exist or does not match the age filter for PyPI package %q", versionToCheck, p.spec.Name), nil
	}

	return false, fmt.Sprintf("version %q does not exist for PyPI package %q", versionToCheck, p.spec.Name), nil
}

package pypi

import (
	"context"
	"errors"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

// Condition checks that a PyPI package version exists.
// It doesn't depend on the versionfilter, which only selects the version returned by the source,
// but it honors the age filter: a version published too recently doesn't pass.
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

	rels, err := p.availableVersions(ctx)
	if err != nil {
		return false, "", err
	}

	status, published := rels.lookup(versionToCheck)
	if published {
		switch status {
		case ageMatching:
			return true, fmt.Sprintf("version %q exists for PyPI package %q", versionToCheck, p.spec.Name), nil
		case ageTooRecent:
			return false, fmt.Sprintf("version %q of PyPI package %q exists but doesn't match the age filter yet", versionToCheck, p.spec.Name), nil
		case ageTooOld:
			return false, fmt.Sprintf("version %q of PyPI package %q exists but is older than the age filter allows", versionToCheck, p.spec.Name), nil
		case ageUndated:
			return false, fmt.Sprintf("version %q of PyPI package %q exists but the registry reports no upload time to check it against the age filter", versionToCheck, p.spec.Name), nil
		}
	}

	return false, fmt.Sprintf("version %q does not exist for PyPI package %q", versionToCheck, p.spec.Name), nil
}

package pypi

import (
	"context"
	"errors"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
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

	if _, _, err := p.availableVersions(ctx); err != nil {
		return false, "", err
	}

	status, published := p.lookupRelease(versionToCheck)
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

// lookupRelease returns the age status of a published version. A version given in its
// normalized form, such as "1.0.0-beta.2" for "1.0b2", is looked up through the releases
// normalizing to it, preferring one matching the age filter. Post releases are left out of
// that lookup: their normalization drops the ".postN" suffix, so "1.0.0" designates the
// base release, not "1.0.0.post1". It must be called after availableVersions.
func (p *Pypi) lookupRelease(ver string) (status releaseAgeStatus, published bool) {
	if status, published := p.releaseStatus[ver]; published {
		return status, true
	}

	if p.versionFilter.Kind == version.PEP440VERSIONKIND {
		return 0, false
	}

	for original, s := range p.releaseStatus {
		if isPostRelease(original) || normalizePEP440(original) != ver {
			continue
		}
		// Statuses are ordered by preference, ageMatching first.
		if !published || s < status {
			status, published = s, true
		}
	}
	return status, published
}

// isPostRelease reports whether a PEP 440 version is a post release, such as "1.0.post1".
func isPostRelease(v string) bool {
	m := pep440PreRelease.FindStringSubmatch(v)
	return m != nil && m[2] == "post"
}

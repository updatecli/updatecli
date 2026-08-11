package version

import (
	"fmt"
	"sort"

	pep440 "github.com/aquasecurity/go-pep440-version"
	"github.com/sirupsen/logrus"
)

// Pep440 filters versions using PEP 440 specifier syntax (e.g. ">=1.0,<2.0").
type Pep440 struct {
	Constraint   string
	versions     []pep440.Version
	FoundVersion Version
}

// Init parses each version string and retains those that are valid PEP 440 versions.
func (p *Pep440) Init(versions []string) error {
	for _, raw := range versions {
		v, err := pep440.Parse(raw)
		if err != nil {
			logrus.Debugf("Skipping %q because %s", raw, err)
			continue
		}
		p.versions = append(p.versions, v)
	}

	if len(p.versions) > 0 {
		return nil
	}

	return ErrNoValidPep440VersionFound
}

// Sort orders versions descending so the newest is first.
func (p *Pep440) Sort() {
	sort.Slice(p.versions, func(i, j int) bool {
		return p.versions[i].Compare(p.versions[j]) > 0
	})
}

// Search returns the highest version satisfying the constraint, or the highest
// version overall when Constraint is empty or "*".
func (p *Pep440) Search(versions []string) error {
	if len(versions) == 0 {
		return ErrNoVersionsFound
	}

	if err := p.Init(versions); err != nil {
		logrus.Error(err)
		return err
	}

	p.Sort()

	if p.Constraint == "" || p.Constraint == "*" {
		// Prefer stable versions over pre-releases.
		for _, v := range p.versions {
			if !v.IsPreRelease() {
				p.FoundVersion.ParsedVersion = v.String()
				p.FoundVersion.OriginalVersion = v.Original()
				return nil
			}
		}
		// Fall back to highest pre-release if no stable version exists.
		p.FoundVersion.ParsedVersion = p.versions[0].String()
		p.FoundVersion.OriginalVersion = p.versions[0].Original()
		return nil
	}

	specifiers, err := pep440.NewSpecifiers(p.Constraint)
	if err != nil {
		// Normalize the xerrors-typed error from the pep440 library to a standard error type.
		return fmt.Errorf("%s", err)
	}

	for _, v := range p.versions {
		if specifiers.Check(v) {
			p.FoundVersion.ParsedVersion = v.String()
			p.FoundVersion.OriginalVersion = v.Original()
			return nil
		}
	}

	return ErrNoVersionFound
}

package osv

import (
	"cmp"
	"fmt"
	"strconv"
	"strings"

	sv "github.com/Masterminds/semver/v3"
	pep440 "github.com/aquasecurity/go-pep440-version"
)

// versionScheme orders the versions of an ecosystem.
type versionScheme interface {
	// Compare returns -1, 0 or +1 when a is lower than, equal to or greater than b.
	Compare(a, b string) (int, error)
	// IsPrerelease returns true when version is a valid pre-release version.
	IsPrerelease(version string) bool
}

// versionSchemeOf returns the version ordering of an ecosystem.
func versionSchemeOf(ecosystem string) (versionScheme, error) {
	switch ecosystem {
	case ecosystemPyPI:
		return pep440Scheme{}, nil
	case ecosystemNuGet:
		return nugetScheme{}, nil
	case ecosystemGo, ecosystemNpm, "crates.io", "Hex", "Pub":
		return semverScheme{}, nil
	}
	return nil, fmt.Errorf("%w %q", ErrUnsupportedEcosystem, ecosystem)
}

// mustCompare compares two versions already known to be valid.
func mustCompare(scheme versionScheme, a, b string) int {
	result, _ := scheme.Compare(a, b)
	return result
}

// semverScheme orders Semantic Versioning versions.
type semverScheme struct{}

func (semverScheme) Compare(a, b string) (int, error) {
	va, err := sv.NewVersion(a)
	if err != nil {
		return 0, err
	}
	vb, err := sv.NewVersion(b)
	if err != nil {
		return 0, err
	}
	return va.Compare(vb), nil
}

func (semverScheme) IsPrerelease(version string) bool {
	v, err := sv.NewVersion(version)
	return err == nil && v.Prerelease() != ""
}

// pep440Scheme orders PyPI versions.
type pep440Scheme struct{}

func (pep440Scheme) Compare(a, b string) (int, error) {
	// Normalize the xerrors-typed errors from the pep440 library to a standard error type.
	va, err := pep440.Parse(a)
	if err != nil {
		return 0, fmt.Errorf("%s", err)
	}
	vb, err := pep440.Parse(b)
	if err != nil {
		return 0, fmt.Errorf("%s", err)
	}
	return va.Compare(vb), nil
}

func (pep440Scheme) IsPrerelease(version string) bool {
	v, err := pep440.Parse(version)
	return err == nil && v.IsPreRelease()
}

// nugetScheme orders NuGet versions, which extend Semantic Versioning with an optional fourth release number.
type nugetScheme struct{}

type nugetVersion struct {
	release    [4]uint64
	prerelease string
}

// parseNuGet parses a NuGet version, ignoring its build metadata.
func parseNuGet(version string) (nugetVersion, error) {
	withoutMetadata, _, _ := strings.Cut(version, "+")
	release, prerelease, _ := strings.Cut(withoutMetadata, "-")

	parts := strings.Split(release, ".")
	if len(parts) > 4 {
		return nugetVersion{}, fmt.Errorf("invalid NuGet version %q", version)
	}

	parsed := nugetVersion{prerelease: prerelease}
	for i, part := range parts {
		n, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return nugetVersion{}, fmt.Errorf("invalid NuGet version %q", version)
		}
		parsed.release[i] = n
	}

	return parsed, nil
}

func (nugetScheme) Compare(a, b string) (int, error) {
	va, err := parseNuGet(a)
	if err != nil {
		return 0, err
	}
	vb, err := parseNuGet(b)
	if err != nil {
		return 0, err
	}

	for i := range va.release {
		if result := cmp.Compare(va.release[i], vb.release[i]); result != 0 {
			return result, nil
		}
	}

	return compareNuGetPrerelease(va.prerelease, vb.prerelease), nil
}

func (nugetScheme) IsPrerelease(version string) bool {
	v, err := parseNuGet(version)
	return err == nil && v.prerelease != ""
}

// compareNuGetPrerelease orders pre-release labels: a release is greater than any pre-release,
// numeric identifiers compare numerically and lower than alphanumeric ones, which compare case-insensitively.
func compareNuGetPrerelease(a, b string) int {
	switch {
	case a == "" && b == "":
		return 0
	case a == "":
		return 1
	case b == "":
		return -1
	}

	identifiersA, identifiersB := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(identifiersA) && i < len(identifiersB); i++ {
		numberA, errA := strconv.ParseUint(identifiersA[i], 10, 64)
		numberB, errB := strconv.ParseUint(identifiersB[i], 10, 64)

		var result int
		switch {
		case errA == nil && errB == nil:
			result = cmp.Compare(numberA, numberB)
		case errA == nil:
			result = -1
		case errB == nil:
			result = 1
		default:
			result = strings.Compare(strings.ToLower(identifiersA[i]), strings.ToLower(identifiersB[i]))
		}

		if result != 0 {
			return result
		}
	}

	return cmp.Compare(len(identifiersA), len(identifiersB))
}

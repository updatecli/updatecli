// Package vprefix retries a changelog lookup with "v" prefixed versions.
// Package registries such as npm and PyPI drop the prefix, while release
// tags on forges often keep it.
package vprefix

import (
	"strings"

	"github.com/updatecli/updatecli/pkg/core/result"
)

// Search runs search with the given versions, then once more with a "v"
// prefix added to each non empty version that lacks one.
func Search(from, to string, search func(from, to string) *result.Changelogs) *result.Changelogs {
	if releases := search(from, to); releases != nil {
		return releases
	}

	vfrom, vto := add(from), add(to)
	if vfrom == from && vto == to {
		return nil
	}

	return search(vfrom, vto)
}

// add prepends "v" to a non empty version that does not already have it.
func add(version string) string {
	if version == "" || strings.HasPrefix(version, "v") {
		return version
	}
	return "v" + version
}

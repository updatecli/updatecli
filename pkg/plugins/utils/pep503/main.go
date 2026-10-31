// Package pep503 implements the Python package name normalization defined by PEP 503
// (https://peps.python.org/pep-0503/#normalized-names).
package pep503

import (
	"regexp"
	"strings"
)

// separators matches the runs of characters PEP 503 normalizes package names on.
var separators = regexp.MustCompile(`[-_.]+`)

// NormalizeName returns the normalized form of a Python package name: lowercase, with
// every run of "-", "_" and "." replaced by a single "-". Two names refer to the same
// package when their normalized forms are equal. PEP 735 normalizes dependency group
// names the same way.
func NormalizeName(name string) string {
	return strings.ToLower(separators.ReplaceAllString(name, "-"))
}

// Package mavenversion orders Maven versions, following the Maven version order specification
// https://maven.apache.org/pom.html#version-order-specification
package mavenversion

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
)

// token is a token of a Maven version, with the separator before it, "." or "-".
type token struct {
	prefix  byte
	value   string
	numeric bool
}

// qualifiers orders the well known qualifiers. Unknown qualifiers come after them, alphabetically.
var qualifiers = map[string]int{
	"alpha":     0,
	"beta":      1,
	"milestone": 2,
	"rc":        3,
	"cr":        3,
	"snapshot":  4,
	"":          5,
	"final":     5,
	"ga":        5,
	"release":   5,
	"sp":        6,
}

// nullQualifiers are the qualifiers equal to a release, trimmed from the end of a version.
var nullQualifiers = []string{"", "final", "ga", "release"}

// shortQualifiers are the qualifiers that can be shortened when directly followed by a number, such as "1-a1".
var shortQualifiers = map[string]string{"a": "alpha", "b": "beta", "m": "milestone"}

// isNull returns true for the tokens trimmed from the end of a version, such as "0" or "ga".
func (t token) isNull() bool {
	if t.numeric {
		return strings.Trim(t.value, "0") == ""
	}
	return slices.Contains(nullQualifiers, t.value)
}

// parse splits a Maven version into tokens between dots, hyphens and transitions between
// digits and characters, which count as a hyphen, then trims the trailing null tokens before
// each hyphen, from the end to the start.
func parse(version string) ([]token, error) {
	if strings.TrimSpace(version) == "" {
		return nil, fmt.Errorf("invalid Maven version %q", version)
	}

	version = strings.ToLower(version)

	var tokens []token
	prefix := byte('.')
	start := 0

	flush := func(end int) {
		value := version[start:end]
		numeric := value != "" && strings.Trim(value, "0123456789") == ""
		if value == "" {
			value, numeric = "0", true
		}
		tokens = append(tokens, token{prefix: prefix, value: value, numeric: numeric})
	}

	for i := 0; i < len(version); i++ {
		c := version[i]
		switch {
		case c == '.' || c == '-':
			flush(i)
			prefix, start = c, i+1
		case i > start && isDigit(c) != isDigit(version[i-1]):
			flush(i)
			if last := &tokens[len(tokens)-1]; isDigit(c) {
				if qualifier, ok := shortQualifiers[last.value]; ok {
					last.value = qualifier
				}
			}
			prefix, start = '-', i
		}
	}
	flush(len(version))

	// Trim the null tokens ending each hyphen group, from the last group to the first one.
	var trimmed []token
	end := len(tokens)
	for i := len(tokens) - 1; i >= 0; i-- {
		if i > 0 && tokens[i].prefix != '-' {
			continue
		}
		group := tokens[i:end]
		for len(group) > 0 && group[len(group)-1].isNull() {
			group = group[:len(group)-1]
		}
		trimmed = append(slices.Clone(group), trimmed...)
		end = i
	}

	return trimmed, nil
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// rank orders tokens with different prefixes: ".qualifier" = "-qualifier" < "-number" < ".number".
func (t token) rank() int {
	switch {
	case !t.numeric:
		return 0
	case t.prefix == '-':
		return 1
	}
	return 2
}

// compareTokens compares two tokens. Qualifiers compare by value whatever their prefix.
func compareTokens(a, b token) int {
	if a.numeric != b.numeric || (a.numeric && a.prefix != b.prefix) {
		return cmp.Compare(a.rank(), b.rank())
	}

	if a.numeric {
		x, y := strings.TrimLeft(a.value, "0"), strings.TrimLeft(b.value, "0")
		if result := cmp.Compare(len(x), len(y)); result != 0 {
			return result
		}
		return strings.Compare(x, y)
	}

	orderA, knownA := qualifiers[a.value]
	orderB, knownB := qualifiers[b.value]
	switch {
	case knownA && knownB:
		return cmp.Compare(orderA, orderB)
	case knownA:
		return -1
	case knownB:
		return 1
	}
	return strings.Compare(a.value, b.value)
}

// nullToken returns the null token padding a version against the other version's token t:
// "0" with the same prefix for a number, the release qualifier "" for a qualifier.
func nullToken(t token) token {
	if t.numeric {
		return token{prefix: t.prefix, value: "0", numeric: true}
	}
	return token{prefix: t.prefix, value: ""}
}

// Compare returns -1, 0 or +1 when a is lower than, equal to or greater than b.
func Compare(a, b string) (int, error) {
	ta, err := parse(a)
	if err != nil {
		return 0, err
	}
	tb, err := parse(b)
	if err != nil {
		return 0, err
	}

	for i := range max(len(ta), len(tb)) {
		var x, y token
		switch {
		case i >= len(ta):
			x, y = nullToken(tb[i]), tb[i]
		case i >= len(tb):
			x, y = ta[i], nullToken(ta[i])
		default:
			x, y = ta[i], tb[i]
		}

		if result := compareTokens(x, y); result != 0 {
			return result, nil
		}
	}

	return 0, nil
}

// IsPrerelease returns true when version holds a qualifier ordered before a release, such as "rc" or "snapshot".
func IsPrerelease(version string) bool {
	tokens, err := parse(version)
	if err != nil {
		return false
	}

	// Pre-release qualifiers, such as "rc" or "snapshot", are ordered before a release.
	for _, t := range tokens {
		if order, known := qualifiers[t.value]; !t.numeric && known && order < qualifiers[""] {
			return true
		}
	}
	return false
}

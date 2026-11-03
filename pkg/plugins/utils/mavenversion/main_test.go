package mavenversion

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompare(t *testing.T) {
	tests := []struct {
		a             string
		b             string
		expected      int
		expectedError bool
	}{
		// Examples from https://maven.apache.org/pom.html#version-order-specification
		{a: "1", b: "1.1", expected: -1},
		{a: "1-snapshot", b: "1", expected: -1},
		{a: "1", b: "1-sp", expected: -1},
		{a: "1-foo2", b: "1-foo10", expected: -1},
		{a: "1-foo", b: "1-1", expected: -1},
		{a: "1-1", b: "1.1", expected: -1},
		{a: "1.ga", b: "1", expected: 0},
		{a: "1-ga", b: "1", expected: 0},
		{a: "1-0", b: "1", expected: 0},
		{a: "1.0", b: "1", expected: 0},
		{a: "1.0.0", b: "1", expected: 0},
		{a: "1.final", b: "1", expected: 0},
		{a: "1-sp", b: "1-ga", expected: 1},
		{a: "1-sp.1", b: "1-ga.1", expected: 1},
		{a: "1-sp-1", b: "1-ga-1", expected: -1},
		{a: "1-a1", b: "1-alpha-1", expected: 0},
		{a: "1.0.0-foo.0.0", b: "1-foo", expected: 0},
		{a: "1.0.0-0.0.0", b: "1", expected: 0},
		{a: "1.foo", b: "1-foo", expected: 0},
		// Padding depends on the type of the other token, so the ordering stays transitive
		{a: "1.foo", b: "1", expected: 1},
		{a: "1.rc", b: "1", expected: -1},
		{a: "9.4.43.v20210629", b: "9.4.43", expected: 1},
		{a: "1-1", b: "1", expected: 1},
		{a: "1.RC1", b: "1-M1", expected: 1},
		// Qualifier order
		{a: "1.0-alpha-1", b: "1.0-beta-1", expected: -1},
		{a: "1.0-beta-2", b: "1.0-milestone-1", expected: -1},
		{a: "1.0-M1", b: "1.0-RC1", expected: -1},
		{a: "1.0-RC1", b: "1.0-CR1", expected: 0},
		{a: "1.0-RC1", b: "1.0-SNAPSHOT", expected: -1},
		{a: "1.0-SNAPSHOT", b: "1.0", expected: -1},
		{a: "1.0", b: "1.0-unknown", expected: -1},
		{a: "1.0-a", b: "1.0-sp", expected: 1},
		// Real versions
		{a: "2.14.1", b: "2.15.0", expected: -1},
		{a: "2.17.0", b: "2.3.1", expected: 1},
		{a: "5.3.20.RELEASE", b: "5.3.20", expected: 0},
		{a: "5.3.19.RELEASE", b: "5.3.20", expected: -1},
		{a: "31.1-jre", b: "32.0.0-jre", expected: -1},
		{a: "31.1-android", b: "31.1-jre", expected: -1},
		{a: "9.4.43.v20210629", b: "9.4.44.v20210927", expected: -1},
		{a: "1.0.0-M2", b: "1.0.0", expected: -1},
		{a: "1.10", b: "1.9", expected: 1},
		{a: "10000000000000000000001", b: "10000000000000000000000", expected: 1},
		{a: "", b: "1.0", expectedError: true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s vs %s", tt.a, tt.b), func(t *testing.T) {
			result, err := Compare(tt.a, tt.b)
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)

			reversed, err := Compare(tt.b, tt.a)
			require.NoError(t, err)
			assert.Equal(t, -tt.expected, reversed)
		})
	}
}

func TestIsPrerelease(t *testing.T) {
	tests := []struct {
		version  string
		expected bool
	}{
		{version: "2.0.0-beta9", expected: true},
		{version: "1.0-SNAPSHOT", expected: true},
		{version: "6.0.0-M1", expected: true},
		{version: "6.0.0.RC2", expected: true},
		{version: "1-a1", expected: true},
		{version: "5.3.20.RELEASE", expected: false},
		{version: "31.1-jre", expected: false},
		{version: "2.17.1", expected: false},
		{version: "", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsPrerelease(tt.version))
		})
	}
}

func TestCompareIsTransitive(t *testing.T) {
	versions := []string{
		"1", "1.0", "1-ga", "1.1", "1-1", "1-0.1", "1.foo", "1-foo", "1-foo1", "1.rc", "1-rc1", "1.RC1",
		"1-M1", "1-alpha", "1-sp", "1-snapshot", "1.0.1-snapshot", "1.1-beta", "9.4.43", "9.4.43.v20210629",
	}

	for _, a := range versions {
		for _, b := range versions {
			for _, c := range versions {
				ab, _ := Compare(a, b)
				bc, _ := Compare(b, c)
				ac, _ := Compare(a, c)
				if ab <= 0 && bc <= 0 {
					assert.LessOrEqual(t, ac, 0, "%s <= %s <= %s", a, b, c)
				}
				if ab == 0 && bc == 0 {
					assert.Equal(t, 0, ac, "%s == %s == %s", a, b, c)
				}
			}
		}
	}
}

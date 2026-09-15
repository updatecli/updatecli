package osv

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNuGetSchemeCompare(t *testing.T) {
	tests := []struct {
		a             string
		b             string
		expected      int
		expectedError bool
	}{
		{a: "1.8.10", b: "1.8.9.1", expected: 1},
		{a: "1.8.6.7", b: "1.8.6.10", expected: -1},
		{a: "1.0", b: "1.0.0.0", expected: 0},
		{a: "1.0.0-rc.1", b: "1.0.0", expected: -1},
		{a: "1.0.0-alpha", b: "1.0.0-beta", expected: -1},
		{a: "1.0.0-rc.2", b: "1.0.0-rc.10", expected: -1},
		{a: "1.0.0-RC.1", b: "1.0.0-rc.1", expected: 0},
		{a: "1.0.0-1", b: "1.0.0-alpha", expected: -1},
		{a: "1.0.0-rc", b: "1.0.0-rc.1", expected: -1},
		{a: "1.0.0+build.1", b: "1.0.0", expected: 0},
		{a: "1.2.3.4.5", b: "1.0.0", expectedError: true},
		{a: "1.0.0", b: "latest", expectedError: true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s vs %s", tt.a, tt.b), func(t *testing.T) {
			result, err := nugetScheme{}.Compare(tt.a, tt.b)
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVersionSchemeIsPrerelease(t *testing.T) {
	tests := []struct {
		scheme   versionScheme
		version  string
		expected bool
	}{
		{scheme: semverScheme{}, version: "6.0.0-rc.1", expected: true},
		{scheme: semverScheme{}, version: "v0.23.0", expected: false},
		{scheme: semverScheme{}, version: "latest", expected: false},
		{scheme: pep440Scheme{}, version: "6.0.0rc1", expected: true},
		{scheme: pep440Scheme{}, version: "6.0.0", expected: false},
		{scheme: pep440Scheme{}, version: "6.0.0.post1", expected: false},
		{scheme: nugetScheme{}, version: "1.0.0-beta", expected: true},
		{scheme: nugetScheme{}, version: "1.8.6.7", expected: false},
		{scheme: nugetScheme{}, version: "1.0.0+build", expected: false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%T %s", tt.scheme, tt.version), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.scheme.IsPrerelease(tt.version))
		})
	}
}

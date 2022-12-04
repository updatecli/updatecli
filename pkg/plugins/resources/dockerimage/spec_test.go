package dockerimage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRegistryEndpoint(t *testing.T) {

	tests := []struct {
		description      string
		image            string
		expectedRegistry string
	}{
		{
			image:            "updatecli/updatecli",
			expectedRegistry: "index.docker.io",
		},
		{
			image:            "ghcr.io/updatecli/updatecli",
			expectedRegistry: "ghcr.io",
		},
	}

	for i := range tests {
		got := sanitizeRegistryEndpoint(tests[i].image)

		if got != tests[i].expectedRegistry {
			t.Errorf("Expected %q but got %q", tests[i].expectedRegistry, got)
		}
	}
}

func TestNewFilterFromValue(t *testing.T) {
	tests := []struct {
		name              string
		expectedTagFilter string
		value             string
	}{
		{
			name:              "Case with latest version",
			expectedTagFilter: `^\d*(\.\d*){2}$`,
			value:             "1.0.0",
		},
		{
			name:              "Case with latest version",
			expectedTagFilter: `^\d*(\.\d*){2}-alpha$`,
			value:             "1.0.0-alpha",
		},
		{
			name:              "Case with jdk",
			expectedTagFilter: `^\d*(\.\d*){1}-jdk11$`,
			value:             "2.235-jdk11",
		},
		{
			name:              "Case with jdk and v prefix",
			expectedTagFilter: `^v\d*(\.\d*){1}-jdk11$`,
			value:             "v2.235-jdk11",
		},
		{
			name:              "Case with jdk",
			expectedTagFilter: `^\d*(\.\d*){1}+jdk11$`,
			value:             "2.235+jdk11",
		},
		{
			name:              "Case with string only",
			expectedTagFilter: "",
			value:             "alpine",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTagFilter, _ := getTagFilterFromValue(tt.value)

			assert.Equal(t, tt.expectedTagFilter, gotTagFilter)
		})
	}
}

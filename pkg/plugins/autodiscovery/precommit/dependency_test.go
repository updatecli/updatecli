package precommit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractVersionFromComment(t *testing.T) {
	tests := []struct {
		name    string
		comment string
		want    string
	}{
		{
			name:    "extracts plain semver",
			comment: "pinned to v2.8.0",
			want:    "v2.8.0",
		},
		{
			name:    "extracts semver with prerelease and build metadata",
			comment: "release v1.2.3-rc.1+build.5",
			want:    "v1.2.3-rc.1+build.5",
		},
		{
			name:    "ignores invalid trailing dash token",
			comment: "v1.2.3- pinned v2.0.0",
			want:    "v2.0.0",
		},
		{
			name:    "returns empty when no semver token found",
			comment: "digest only",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, extractVersionFromComment(tt.comment))
		})
	}
}

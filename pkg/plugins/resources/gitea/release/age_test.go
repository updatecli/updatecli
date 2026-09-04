package release

import (
	"testing"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/stretchr/testify/assert"
)

func TestReleaseDate(t *testing.T) {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	published := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		release scm.Release
		want    time.Time
	}{
		{
			name:    "publication date is used when set",
			release: scm.Release{Created: created, Published: published},
			want:    published,
		},
		{
			name:    "creation date is used when the release has no publication date",
			release: scm.Release{Created: created},
			want:    created,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, releaseDate(&tt.release))
		})
	}
}

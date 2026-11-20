package age

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestSearch(t *testing.T) {
	tests := []struct {
		name      string
		pattern   string
		published []string
		kept      []string
		want      string
		wantHold  bool
		wantError bool
	}{
		{
			name:      "a kept version matches",
			pattern:   ">=1.5.0",
			published: []string{"1.4.0", "1.5.0", "1.6.0"},
			kept:      []string{"1.4.0", "1.5.0"},
			want:      "1.5.0",
		},
		{
			name:      "every matching version is too recent",
			pattern:   ">=1.5.0",
			published: []string{"1.4.0", "1.5.0", "1.6.0"},
			kept:      []string{"1.4.0"},
			wantHold:  true,
		},
		{
			name:      "no published version matches",
			pattern:   ">=2.0.0",
			published: []string{"1.4.0", "1.5.0"},
			kept:      []string{"1.4.0"},
			wantError: true,
		},
		{
			name:      "nothing discarded by age",
			pattern:   ">=2.0.0",
			published: []string{"1.4.0", "1.5.0"},
			kept:      []string{"1.4.0", "1.5.0"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := version.Filter{Kind: version.SEMVERVERSIONKIND, Pattern: tt.pattern}

			found, err := Search(&filter, tt.published, tt.kept)

			switch {
			case tt.wantHold:
				require.Error(t, err)
				assert.True(t, errors.Is(err, ErrNoVersionMatchingAge), "got %v", err)
			case tt.wantError:
				require.Error(t, err)
				assert.False(t, errors.Is(err, ErrNoVersionMatchingAge), "a missing version must not be reported as an age hold")
			default:
				require.NoError(t, err)
				assert.Equal(t, tt.want, found.GetVersion())
			}
		})
	}
}

func TestHeldBack(t *testing.T) {
	filter := version.Filter{Kind: "semver", Pattern: ">=1.5.0"}

	assert.True(t, HeldBack(&filter, []string{"1.4.0", "1.5.1"}), "a discarded version matches the filter")
	assert.False(t, HeldBack(&filter, []string{"1.4.0"}), "no discarded version matches the filter")
	assert.False(t, HeldBack(&filter, nil), "nothing discarded")
}

func TestMatchesUndated(t *testing.T) {
	filter := version.Filter{Kind: "semver", Pattern: ">=1.5.0"}
	published := []string{"1.4.0", "1.5.0"}

	assert.True(t, MatchesUndated(&filter, published, []string{"1.4.0"}), "a version matching the filter has no release date")
	assert.False(t, MatchesUndated(&filter, published, published), "every version has a release date")
	assert.False(t, MatchesUndated(&filter, published, []string{"1.5.0"}), "the version without release date doesn't match the filter")
}

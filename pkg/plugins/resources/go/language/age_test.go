package language

import (
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestSearchDatedVersions(t *testing.T) {
	daysAgo := func(n int) time.Time {
		return time.Now().Add(-time.Duration(n) * 24 * time.Hour)
	}

	tags := func() []datedVersion {
		return []datedVersion{
			{version: semver.MustParse("1.25.4"), releaseDate: daysAgo(2)},
			{version: semver.MustParse("1.24.0"), releaseDate: daysAgo(90)},
			{version: semver.MustParse("1.25.3"), releaseDate: daysAgo(30)},
		}
	}

	tests := []struct {
		name            string
		pattern         string
		releaseAge      age.Spec
		expectedVersion string
		expectedHold    bool
		expectedError   bool
	}{
		{
			name:            "the newest version out of cooldown matches",
			pattern:         "1.25.x",
			releaseAge:      age.Spec{Minimum: "7d"},
			expectedVersion: "1.25.3",
		},
		{
			// The autodiscovery bounds the filter at the version in go.mod, so while that
			// version is too recent, only older versions are left
			name:         "every version matching the filter is in cooldown",
			pattern:      ">=1.25.4, 1.25.x",
			releaseAge:   age.Spec{Minimum: "7d"},
			expectedHold: true,
		},
		{
			name:          "a version filter matching no version still fails",
			pattern:       ">=1.26.0",
			releaseAge:    age.Spec{Minimum: "7d"},
			expectedError: true,
		},
		{
			name:          "a version filter matching no version still fails while every version is in cooldown",
			pattern:       ">=1.26.0",
			releaseAge:    age.Spec{Minimum: "1y"},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, err := New(Spec{
				VersionFilter: version.Filter{Kind: "semver", Pattern: tt.pattern},
				Age:           tt.releaseAge,
			})
			require.NoError(t, err)

			_, err = l.searchDatedVersions(tags())

			switch {
			case tt.expectedHold:
				require.ErrorIs(t, err, age.ErrNoVersionMatchingAge)
			case tt.expectedError:
				require.Error(t, err)
				assert.NotErrorIs(t, err, age.ErrNoVersionMatchingAge)
			default:
				require.NoError(t, err)
				assert.Equal(t, tt.expectedVersion, l.Version.GetVersion())
			}
		})
	}
}

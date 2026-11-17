package version

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearch(t *testing.T) {
	tests := []struct {
		name     string
		filter   Filter
		versions []string
		want     Version
		wantErr  error
	}{
		{
			name: "Passing case with filter 'latest' and pattern 'latest'",
			filter: Filter{
				Kind:    LATESTVERSIONKIND,
				Pattern: LATESTVERSIONKIND,
			},
			versions: []string{"1.0", "2.0", "3.0"},
			want: Version{
				ParsedVersion:   "3.0",
				OriginalVersion: "3.0",
			},
		},
		{
			name: "Passing case with filter 'latest' but custom specified pattern",
			filter: Filter{
				Kind:    LATESTVERSIONKIND,
				Pattern: "2.0",
			},
			versions: []string{"1.0", "2.0", "3.0"},
			want: Version{
				ParsedVersion:   "2.0",
				OriginalVersion: "2.0",
			},
		},
		{
			name: "Passing case with filter semver and pattern",
			filter: Filter{
				Kind:    SEMVERVERSIONKIND,
				Pattern: "~2",
			},
			versions: []string{"1.0", "2.0", "3.0"},
			want: Version{
				ParsedVersion:   "2.0.0",
				OriginalVersion: "2.0",
			},
		},
		{
			name: "Passing case with filter semver but no pattern",
			filter: Filter{
				Kind: SEMVERVERSIONKIND,
			},
			versions: []string{"1.0", "2.0", "3.0"},
			want: Version{
				ParsedVersion:   "3.0.0",
				OriginalVersion: "3.0",
			},
		},
		{
			name: "Passing case with filter semver and prerelease",
			filter: Filter{
				Kind:    SEMVERVERSIONKIND,
				Pattern: "3.x.x-0",
			},
			versions: []string{"1.0.0", "2.0.0", "3.0.0-rc1"},
			want: Version{
				ParsedVersion:   "3.0.0-rc1",
				OriginalVersion: "3.0.0-rc1",
			},
		},
		{
			name: "Passing case with filter semver, prerelease, and minor update",
			filter: Filter{
				Kind:    SEMVERVERSIONKIND,
				Pattern: "1.x.x-0",
			},
			versions: []string{"1.1.0", "1.2.0", "1.3.0-rc1", "2.0.0"},
			want: Version{
				ParsedVersion:   "1.3.0-rc1",
				OriginalVersion: "1.3.0-rc1",
			},
		},
		{
			name: "Passing case with filter semver and minor update",
			filter: Filter{
				Kind:    SEMVERVERSIONKIND,
				Pattern: "1.x.x",
			},
			versions: []string{"1.1.0", "1.2.0", "1.3.0-rc1", "2.0.0"},
			want: Version{
				ParsedVersion:   "1.2.0",
				OriginalVersion: "1.2.0",
			},
		},
		{
			name: "Failing case with no semver (+pattern) found",
			filter: Filter{
				Kind:    SEMVERVERSIONKIND,
				Pattern: "~2",
			},
			versions: []string{"updatecli-1.0", "updatecli-2.0", "updatecli-3.0"},
			want:     Version{},
			wantErr:  errors.New("no valid semantic version found"),
		},
		{
			name: "Passing case with regexp filter and pattern",
			filter: Filter{
				Kind:    REGEXVERSIONKIND,
				Pattern: "^updatecli-2.(\\d*)$",
			},
			versions: []string{"updatecli-1.0", "updatecli-2.0", "updatecli-3.0"},
			want: Version{
				ParsedVersion:   "updatecli-2.0",
				OriginalVersion: "updatecli-2.0",
			},
		},
		{
			name: "Passing case with regexp filter but no pattern",
			filter: Filter{
				Kind: REGEXVERSIONKIND,
			},
			versions: []string{"updatecli-1.0", "updatecli-2.0", "updatecli-3.0"},
			want: Version{
				ParsedVersion:   "updatecli-3.0",
				OriginalVersion: "updatecli-3.0",
			},
		},
		{
			name: "Failing case with regexp filter (+pattern)",
			filter: Filter{
				Kind:    REGEXVERSIONKIND,
				Pattern: "^updatecli-4.(\\d*)$",
			},
			versions: []string{"updatecli-1.0", "updatecli-2.0", "updatecli-3.0"},
			want:     Version{},
			wantErr:  &ErrNoVersionFoundForPattern{Pattern: "^updatecli-4.(\\d*)$"},
		},
		{
			name: "Passing case with regex/semver",
			filter: Filter{
				Kind:  REGEXSEMVERVERSIONKIND,
				Regex: "^updatecli-(\\d*\\.\\d*\\.\\d*)$",
			},
			versions: []string{"updatecli-1.0.0", "updatecli-2.0.0", "updatecli-1.1.0"},
			want: Version{
				ParsedVersion:   "2.0.0",
				OriginalVersion: "updatecli-2.0.0",
			},
		},
		{
			name: "Passing case with regex/semver",
			filter: Filter{
				Kind:  REGEXSEMVERVERSIONKIND,
				Regex: "^updatecli-\\d*\\.\\d*\\.\\d*$",
			},
			versions: []string{"updatecli-1.0.0", "updatecli-2.0.0", "updatecli-1.1.0"},
			want:     Version{},
			wantErr:  errors.New("versions list empty"),
		},
		{
			name: "Empty case with regex/semver",
			filter: Filter{
				Kind: REGEXSEMVERVERSIONKIND,
			},
			versions: []string{"updatecli-1.0.0", "updatecli-2.0.0", "updatecli-1.1.0"},
			want:     Version{},
			wantErr:  errors.New("versions list empty"),
		},
		{
			name: "Passing case with regex/time",
			filter: Filter{
				Kind:    REGEXTIMEVERSIONKIND,
				Regex:   `^updatecli-(\d*)$`,
				Pattern: "20060201",
			},
			versions: []string{"updatecli-20232103", "updatecli-20232205", "updatecli-20200101"},
			want: Version{
				ParsedVersion:   "updatecli-20232205",
				OriginalVersion: "updatecli-20232205",
			},
		},
		{
			name: "Passing case with time",
			filter: Filter{
				Kind:    TIMEVERSIONKIND,
				Pattern: "20060201",
			},
			versions: []string{"20232103", "20232205", "20200101"},
			want: Version{
				ParsedVersion:   "20232205",
				OriginalVersion: "20232205",
			},
		},
		{
			name: "Passing case with lex",
			filter: Filter{
				Kind: LEXVERSIONKIND,
			},
			versions: []string{"main-20241226-57-2242a78", "main-20240910-43-3705031", "main-20240916-49-bc46fdf"},
			want: Version{
				ParsedVersion:   "main-20241226-57-2242a78",
				OriginalVersion: "main-20241226-57-2242a78",
			},
		},
		{
			name: "Passing case with lex",
			filter: Filter{
				Kind: LEXVERSIONKIND,
			},
			versions: []string{"v1.1", "v1.10", "v1.2"},
			want: Version{
				ParsedVersion:   "v1.2",
				OriginalVersion: "v1.2",
			},
		},
		{
			name: "Passing case with regex/semver and replaceAll (curl example)",
			filter: Filter{
				Kind:  REGEXSEMVERVERSIONKIND,
				Regex: "curl-(\\d+\\.\\d+\\.\\d+)",
				ReplaceAll: ReplaceAll{
					Pattern:     "_",
					Replacement: ".",
				},
				Pattern: "8.15.x",
			},
			versions: []string{"curl-8_15_0", "curl-8_14_0", "curl-8_16_0"},
			want: Version{
				ParsedVersion:   "8.15.0",
				OriginalVersion: "curl-8_15_0",
			},
		},
		{
			name: "Passing case with regex/semver and replaceAll with capture groups",
			filter: Filter{
				Kind:  REGEXSEMVERVERSIONKIND,
				Regex: "v(\\d+\\.\\d+\\.\\d+)",
				ReplaceAll: ReplaceAll{
					Pattern:     "([0-9]+)_([0-9]+)_([0-9]+)",
					Replacement: "$1.$2.$3",
				},
				Pattern: ">=1.0.0",
			},
			versions: []string{"v1_2_3", "v2_0_0", "v0_9_9"},
			want: Version{
				ParsedVersion:   "2.0.0",
				OriginalVersion: "v2_0_0",
			},
		},
		{
			name: "Passing case with regex/time and replaceAll",
			filter: Filter{
				Kind:    REGEXTIMEVERSIONKIND,
				Regex:   `^updatecli(\d{8})$`,
				Pattern: "20060201",
				ReplaceAll: ReplaceAll{
					Pattern:     "-",
					Replacement: "",
				},
			},
			versions: []string{"updatecli-2023-02-01", "updatecli-2023-03-05", "updatecli-2023-01-15"},
			want: Version{
				ParsedVersion:   "updatecli-2023-03-05",
				OriginalVersion: "updatecli-2023-03-05",
			},
		},
		{
			name: "Passing case with regex and replaceAll",
			filter: Filter{
				Kind:    REGEXVERSIONKIND,
				Pattern: "^updatecli-2\\.(\\d+)$",
				ReplaceAll: ReplaceAll{
					Pattern:     "_",
					Replacement: ".",
				},
			},
			versions: []string{"updatecli-2_0", "updatecli-2_1", "updatecli-3_0"},
			want: Version{
				ParsedVersion:   "updatecli-2.1",
				OriginalVersion: "updatecli-2_1",
			},
		},
		{
			name: "Passing case with regex/semver and replaceAll (no match in replaceAll)",
			filter: Filter{
				Kind:  REGEXSEMVERVERSIONKIND,
				Regex: "^updatecli-(\\d+\\.\\d+\\.\\d+)$",
				ReplaceAll: ReplaceAll{
					Pattern:     "X",
					Replacement: ".",
				},
			},
			versions: []string{"updatecli-1.0.0", "updatecli-2.0.0", "updatecli-1.1.0"},
			want: Version{
				ParsedVersion:   "2.0.0",
				OriginalVersion: "updatecli-2.0.0",
			},
		},
		{
			name: "Passing case with regex/semver and replaceAll (empty replacement)",
			filter: Filter{
				Kind:  REGEXSEMVERVERSIONKIND,
				Regex: "^updatecli(\\d+\\.\\d+\\.\\d+)$",
				ReplaceAll: ReplaceAll{
					Pattern:     "-",
					Replacement: "",
				},
			},
			versions: []string{"updatecli-1.0.0", "updatecli-2.0.0", "updatecli-1.1.0"},
			want: Version{
				ParsedVersion:   "2.0.0",
				OriginalVersion: "updatecli-2.0.0",
			},
		},
		{
			name: "Passing case with filter pep440 and no pattern",
			filter: Filter{
				Kind: PEP440VERSIONKIND,
			},
			versions: []string{"1.0", "2.0", "3.0"},
			want: Version{
				ParsedVersion:   "3.0",
				OriginalVersion: "3.0",
			},
		},
		{
			name: "Passing case with filter pep440 and constraint",
			filter: Filter{
				Kind:    PEP440VERSIONKIND,
				Pattern: ">=1.0,<3.0",
			},
			versions: []string{"1.0", "2.0", "3.0"},
			want: Version{
				ParsedVersion:   "2.0",
				OriginalVersion: "2.0",
			},
		},
		{
			name: "Failing case with no pep440 version found",
			filter: Filter{
				Kind:    PEP440VERSIONKIND,
				Pattern: ">=4.0",
			},
			versions: []string{"1.0", "2.0", "3.0"},
			want:     Version{},
			wantErr:  errors.New("no version found"),
		},
		{
			name: "Passing case with pep440 pre-release versions",
			filter: Filter{
				Kind:    PEP440VERSIONKIND,
				Pattern: ">=1.0",
			},
			versions: []string{"1.0a1", "1.0b1", "1.0rc1", "1.0"},
			want: Version{
				ParsedVersion:   "1.0",
				OriginalVersion: "1.0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.filter.Search(tt.versions)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		filter  Filter
		wantErr error
	}{
		{
			name: "Valid semver filter",
			filter: Filter{
				Kind:    SEMVERVERSIONKIND,
				Pattern: "~2",
			},
			wantErr: nil,
		},
		{
			name: "Valid regex filter",
			filter: Filter{
				Kind:    REGEXVERSIONKIND,
				Pattern: "~2",
			},
			wantErr: nil,
		},
		{
			name: "Invalid kind of filter",
			filter: Filter{
				Kind:    "noExist",
				Pattern: "~2",
			},
			wantErr: &ErrUnsupportedVersionKind{Kind: "noExist"},
		},
		{
			name: "Valid filter with replaceAll",
			filter: Filter{
				Kind:  REGEXSEMVERVERSIONKIND,
				Regex: "^updatecli-(\\d+\\.\\d+\\.\\d+)$",
				ReplaceAll: ReplaceAll{
					Pattern:     "_",
					Replacement: ".",
				},
			},
			wantErr: nil,
		},
		{
			name: "Invalid replaceAll pattern",
			filter: Filter{
				Kind:  REGEXSEMVERVERSIONKIND,
				Regex: "^updatecli-(\\d+\\.\\d+\\.\\d+)$",
				ReplaceAll: ReplaceAll{
					Pattern:     "[invalid",
					Replacement: ".",
				},
			},
			wantErr: errors.New("replaceAll pattern validation error"), // Marker error to indicate we expect an error
		},
		{
			name: "Valid replaceAll with empty replacement",
			filter: Filter{
				Kind:  REGEXSEMVERVERSIONKIND,
				Regex: "^updatecli-(\\d+\\.\\d+\\.\\d+)$",
				ReplaceAll: ReplaceAll{
					Pattern:     "_",
					Replacement: "",
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()

			if tt.wantErr != nil {
				require.Error(t, err)
				if tt.name == "Invalid replaceAll pattern" {
					// For invalid replaceAll pattern, check that error contains expected message
					assert.Contains(t, err.Error(), "invalid replaceAll pattern")
				} else {
					assert.Equal(t, tt.wantErr, err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewFilter(t *testing.T) {
	tests := []struct {
		name    string
		filter  Filter
		want    Filter
		wantErr bool
	}{
		{
			name: "Case with latest version",
			filter: Filter{
				Kind:    "latest",
				Pattern: "latest",
			},
			want: Filter{
				Kind:    "latest",
				Pattern: "latest",
			},
		},
		{
			name:   "Case with empty arguments",
			filter: Filter{},
			want: Filter{
				Kind:    LATESTVERSIONKIND,
				Pattern: "latest",
			},
		},
		{
			name: "Case with empty pattern for semver",
			filter: Filter{
				Kind:    SEMVERVERSIONKIND,
				Pattern: "",
			},
			want: Filter{
				Kind:    SEMVERVERSIONKIND,
				Pattern: "*",
			},
		},
		{
			name: "Case with empty pattern for regexp",
			filter: Filter{
				Kind:    REGEXVERSIONKIND,
				Pattern: "",
			},
			want: Filter{
				Kind:    REGEXVERSIONKIND,
				Pattern: ".*",
			},
		},
		{
			name: "Case with empty pattern for pep440",
			filter: Filter{
				Kind:    PEP440VERSIONKIND,
				Pattern: "",
			},
			want: Filter{
				Kind:    PEP440VERSIONKIND,
				Pattern: "*",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.filter.Init()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGreaterThanPattern(t *testing.T) {
	tests := []struct {
		name    string
		filter  Filter
		version string
		want    string
		wantErr error
	}{
		{
			name: "Latest version kind",
			filter: Filter{
				Kind:    LATESTVERSIONKIND,
				Pattern: LATESTVERSIONKIND,
			},
			version: "3.0", want: "latest",
		},
		{
			name: "Regex version kind",
			filter: Filter{
				Kind:    REGEXVERSIONKIND,
				Pattern: "^3.*",
			},
			version: "3.0", want: "^3.*",
		},
		{
			name: "Major semver pattern",
			filter: Filter{
				Kind:    SEMVERVERSIONKIND,
				Pattern: "major",
			},
			version: "3.0", want: ">=3",
		},
		{
			name: "Major semver pattern with prerelease",
			filter: Filter{
				Kind:    SEMVERVERSIONKIND,
				Pattern: "major",
			},
			version: "3.0.0-rc1", want: ">=3.x.x-0",
		},
		{
			name: "Minor semver pattern",
			filter: Filter{
				Kind:    SEMVERVERSIONKIND,
				Pattern: "minor",
			},
			version: "3.0", want: "3.x",
		},
		{
			name: "Minor semver pattern with prerelease",
			filter: Filter{
				Kind:    SEMVERVERSIONKIND,
				Pattern: "minor",
			},
			version: "3.0.0-rc1", want: "3.x.x-0",
		},
		{
			name: "Minor semver only pattern",
			filter: Filter{
				Kind:    SEMVERVERSIONKIND,
				Pattern: "minoronly",
			},
			version: "3.1", want: "3.1 || >3.1 < 4",
		},
		{
			name: "Major semver only pattern",
			filter: Filter{
				Kind:    SEMVERVERSIONKIND,
				Pattern: "majoronly",
			},
			version: "3.1", want: "3.1 || >3",
		},
		{
			name: "Major semver only pattern with semver",
			filter: Filter{
				Kind:    SEMVERVERSIONKIND,
				Pattern: "majoronly",
			},
			version: "3.1.0-rc1", want: "3.1.0-rc1 || >3.1.0-rc1",
		},
		{
			name: "Patch semver pattern",
			filter: Filter{
				Kind:    SEMVERVERSIONKIND,
				Pattern: "patch",
			},
			version: "3.0", want: "3.0.x",
		},
		{
			name: "Prerelease semver pattern",
			filter: Filter{
				Kind:    SEMVERVERSIONKIND,
				Pattern: "prerelease",
			},
			version: "3.0", want: ">=3.0.0-0 <= 3.0.0",
		},
		{
			name: "Version constraint semver pattern",
			filter: Filter{
				Kind:    SEMVERVERSIONKIND,
				Pattern: "*",
			},
			version: "1.0 - 2.0", want: "1.0 - 2.0",
		},
		{
			name: "Wrong Version constraint semver pattern",
			filter: Filter{
				Kind:    SEMVERVERSIONKIND,
				Pattern: "*",
			},
			version: "v0.0.0-20220606043923-3cf50f8a0a29",
			want:    ">=0.0.0-20220606043923-3cf50f8a0a29",
		},
		{
			name: "Wrong Semver Version",
			filter: Filter{
				Kind:    SEMVERVERSIONKIND,
				Pattern: "*",
			},
			version: "v0.0.0_20220606043923-3cf50f8a0a29",
			want:    "",
			wantErr: &ErrIncorrectSemVerConstraint{SemVerConstraint: "v0.0.0_20220606043923-3cf50f8a0a29"},
		},
		{
			name: "Wrong Semver Constraint",
			filter: Filter{
				Kind:    SEMVERVERSIONKIND,
				Pattern: "*",
			},
			version: "1.0 - 2.0 !!!",
			want:    "",
			wantErr: &ErrIncorrectSemVerConstraint{SemVerConstraint: "1.0 - 2.0 !!!"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.filter.GreaterThanPattern(tt.version)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGreaterThanPatternWithoutDowngrade(t *testing.T) {
	tests := []struct {
		pattern string
		version string
		want    string
	}{
		{pattern: "*", version: "1.5.3", want: ">=1.5.3"},
		{pattern: "patch", version: "3.0", want: ">=3.0.0, 3.0.x"},
		{pattern: "minor", version: "3.0", want: ">=3.0.0, 3.x"},
		{pattern: "minor", version: "3.0.0-rc1", want: ">=3.0.0-rc1, 3.x.x-0"},
		{pattern: "major", version: "3.0", want: ">=3.0.0"},
		{pattern: "major", version: "3.0.0-rc1", want: ">=3.0.0-rc1, >=3.x.x-0"},
		{pattern: "prerelease", version: "1.2.3", want: ">=1.2.3, >=1.2.3-0 <= 1.2.3"},
		// A custom constraint accepting older versions is bounded at the current version
		{pattern: "~1.5", version: "1.5.3", want: ">=1.5.3, ~1.5"},
		// Each group starting below the current version is bounded
		{pattern: "^1.0.0 || ^2.0.0", version: "1.5.3", want: ">=1.5.3, ^1.0.0 || ^2.0.0"},
		// A custom constraint already starting above the current version is unchanged
		{pattern: ">=2.0.0", version: "1.5.3", want: ">=2.0.0"},
		// A custom constraint with a current version that isn't semver is unchanged
		{pattern: "~1.5", version: "not-a-version", want: "~1.5"},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+" "+tt.version, func(t *testing.T) {
			filter := Filter{Kind: SEMVERVERSIONKIND, Pattern: tt.pattern}
			got, err := filter.GreaterThanPatternWithoutDowngrade(tt.version)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)

			// Without an age filter, the pattern is left as it was
			if tt.pattern == "~1.5" && tt.version == "1.5.3" {
				legacy, err := filter.GreaterThanPattern(tt.version)
				require.NoError(t, err)
				assert.Equal(t, "~1.5", legacy)
			}
		})
	}
}

// TestGreaterThanPatternNeverDowngrades checks that the generated semver patterns never pick a
// version lower than the current one, such as when an age filter drops the current version
// and every newer one.
func TestGreaterThanPatternNeverDowngrades(t *testing.T) {
	for _, pattern := range []string{"*", "patch", "minor", "minoronly", "major", "majoronly", "prerelease"} {
		t.Run(pattern, func(t *testing.T) {
			filter := Filter{Kind: SEMVERVERSIONKIND, Pattern: pattern}
			greaterThanPattern, err := filter.GreaterThanPatternWithoutDowngrade("1.5.3")
			require.NoError(t, err)

			bounded := Filter{Kind: SEMVERVERSIONKIND, Pattern: greaterThanPattern}

			// Only older versions are left, so nothing must match.
			_, err = bounded.Search([]string{"1.4.0", "1.5.0", "1.5.2"})
			assert.Error(t, err, "pattern %q must not accept a version lower than 1.5.3", greaterThanPattern)

			// The current version is still accepted, so an update never goes backwards.
			found, err := bounded.Search([]string{"1.4.0", "1.5.2", "1.5.3"})
			require.NoError(t, err)
			assert.Equal(t, "1.5.3", found.GetVersion())
		})
	}
}

func TestGreaterThanPatternAllowingDowngrade(t *testing.T) {
	tests := []struct {
		pattern string
		version string
		want    string
	}{
		{pattern: "*", version: "1.5.3", want: "*"},
		// "*" would reject prereleases, replacing a prerelease in use by an older stable release
		{pattern: "*", version: "2.0.0-rc.2", want: ">=0.0.0-0"},
		{pattern: "patch", version: "1.5.3", want: "1.5.x"},
		{pattern: "patch", version: "1.5.3-rc1", want: "1.5.x-0"},
		{pattern: "minor", version: "1.5.3", want: "1.x"},
		{pattern: "minor", version: "1.5.3-rc1", want: "1.x.x-0"},
		{pattern: "major", version: "1.5.3", want: ">=1"},
		{pattern: "major", version: "1.5.3-rc1", want: ">=1.x.x-0"},
		// Patterns without any lower bound to remove are the same as GreaterThanPatternWithoutDowngrade
		{pattern: "minoronly", version: "1.5.3", want: "1.5.3 || >1.5 < 2"},
		// A custom constraint is returned as is
		{pattern: "~1.2", version: "1.5.3", want: "~1.2"},
		{pattern: "~1.5", version: "1.5.3", want: "~1.5"},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+" "+tt.version, func(t *testing.T) {
			filter := Filter{Kind: SEMVERVERSIONKIND, Pattern: tt.pattern}
			got, err := filter.GreaterThanPatternAllowingDowngrade(tt.version)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	// Once the newer versions are filtered out, an older one is accepted
	for _, pattern := range []string{"*", "patch", "minor", "major"} {
		t.Run(pattern+" falls back to an older version", func(t *testing.T) {
			filter := Filter{Kind: SEMVERVERSIONKIND, Pattern: pattern}
			got, err := filter.GreaterThanPatternAllowingDowngrade("1.5.3")
			require.NoError(t, err)

			found, err := (&Filter{Kind: SEMVERVERSIONKIND, Pattern: got}).Search([]string{"1.4.0", "1.5.0", "1.5.2"})
			require.NoError(t, err)
			assert.Equal(t, "1.5.2", found.GetVersion())
		})
	}
}

func TestBoundPattern(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		lowerBound string
		want       string
	}{
		{name: "any version", pattern: "*", lowerBound: "1.2.5", want: ">=1.2.5"},
		{name: "empty pattern", pattern: "", lowerBound: "1.2.5", want: ">=1.2.5"},
		{name: "caret range", pattern: "^1.2.0", lowerBound: "1.2.5", want: ">=1.2.5, ^1.2.0"},
		{name: "each group below the lower bound", pattern: "^1.0.0 || ^2.0.0", lowerBound: "1.2.5", want: ">=1.2.5, ^1.0.0 || ^2.0.0"},
		{name: "group already bounded", pattern: ">=1.2.5, 1.x", lowerBound: "1.2.5", want: ">=1.2.5, 1.x"},
		{name: "v prefix", pattern: "~1.2", lowerBound: "v1.2.5", want: ">=1.2.5, ~1.2"},
		// The bound doesn't make a group accept prereleases, so a prerelease is bounded at its release
		{name: "prerelease", pattern: "^1.2.0", lowerBound: "1.5.0-beta.1", want: ">=1.5.0, ^1.2.0"},
		// Hyphens in build metadata aren't prereleases
		{name: "prerelease with build metadata", pattern: "^1.2.0+2026-01", lowerBound: "1.5.0-beta.1", want: ">=1.5.0, ^1.2.0+2026-01"},
		{name: "prerelease with any version", pattern: "*", lowerBound: "1.5.0-beta.1", want: ">=1.5.0"},
		{name: "prerelease below the pattern", pattern: "~1.2", lowerBound: "v0.0.0-20230215024106-420ad0987b9b", want: "~1.2"},
		{name: "prerelease with a group accepting prereleases", pattern: "1.5.x-0 || ^2.0.0", lowerBound: "1.5.0-beta.1", want: ">=1.5.0-beta.1, 1.5.x-0 || ^2.0.0"},
		{name: "prerelease with any version or prerelease", pattern: ">=0.0.0-0", lowerBound: "1.5.0-beta.1", want: ">=1.5.0-beta.1"},
		// The release bound must not drop the prereleases the pattern accepts
		{name: "release with any version or prerelease", pattern: ">=0.0.0-0", lowerBound: "1.5.0", want: ">=1.5.0, >=0.0.0-0"},
		{name: "prerelease already bounded", pattern: "^2.0.0-beta.3", lowerBound: "2.0.0-beta.3", want: "^2.0.0-beta.3"},
		{name: "upper bound at the lower bound", pattern: "<1.2.5", lowerBound: "1.2.5", want: ">=1.2.5, <1.2.5"},
		{name: "inclusive upper bound at the lower bound", pattern: "<=1.2.5", lowerBound: "1.2.5", want: ">=1.2.5, <=1.2.5"},
		{name: "exclusion of the lower bound", pattern: "!=1.2.5", lowerBound: "1.2.5", want: ">=1.2.5, !=1.2.5"},
		{name: "hyphen range ending at the lower bound", pattern: "1.0.0 - 1.2.5", lowerBound: "1.2.5", want: ">=1.2.5, 1.0.0 - 1.2.5"},
		{name: "hyphen range starting at the lower bound", pattern: "1.2.5 - 2.0.0", lowerBound: "1.2.5", want: "1.2.5 - 2.0.0"},
		{name: "lower bound inside a bigger version", pattern: "<11.2.5", lowerBound: "1.2.5", want: ">=1.2.5, <11.2.5"},
		{name: "operator followed by a space", pattern: ">= 1.2.5", lowerBound: "1.2.5", want: ">= 1.2.5"},
		{name: "higher lower bound", pattern: ">=1.3.0", lowerBound: "1.2.5", want: ">=1.3.0"},
		{name: "exact version", pattern: "1.2.5", lowerBound: "1.2.5", want: "1.2.5"},
		{name: "not a constraint", pattern: "latest", lowerBound: "1.2.5", want: "latest"},
		{name: "not a version", pattern: "^1.2.0", lowerBound: "next", want: "^1.2.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, BoundPattern(tt.pattern, tt.lowerBound))
		})
	}
}

func TestBoundPatternAcceptingPrereleases(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		lowerBound string
		want       string
	}{
		{name: "pseudo version", pattern: "~1.2", lowerBound: "v1.2.4-0.20230215024106-420ad0987b9b", want: ">=1.2.4-0.20230215024106-420ad0987b9b, ~1.2"},
		// The prerelease bound also lets the group accept prereleases
		{name: "zero pseudo version below the pattern", pattern: "~1.2", lowerBound: "v0.0.0-20230215024106-420ad0987b9b", want: ">=0.0.0-20230215024106-420ad0987b9b, ~1.2"},
		{name: "pseudo version with any version", pattern: "*", lowerBound: "v0.0.0-20230215024106-420ad0987b9b", want: ">=0.0.0-20230215024106-420ad0987b9b"},
		{name: "pseudo version with any version or prerelease", pattern: ">=0.0.0-0", lowerBound: "v0.0.0-20230215024106-420ad0987b9b", want: ">=0.0.0-20230215024106-420ad0987b9b"},
		{name: "release", pattern: "^1.2.0", lowerBound: "1.2.5", want: ">=1.2.5, ^1.2.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, BoundPatternAcceptingPrereleases(tt.pattern, tt.lowerBound))
		})
	}
}

// TestBoundPatternKeepsStableGroupsStable checks that bounding a stable range at a prerelease,
// such as a prerelease in a lock file, doesn't make the range accept prereleases.
func TestBoundPatternKeepsStableGroupsStable(t *testing.T) {
	filter := Filter{Kind: SEMVERVERSIONKIND, Pattern: BoundPattern("^1.2.0", "1.5.0-beta.1")}

	found, err := filter.Search([]string{"1.4.0", "1.5.0-beta.1", "1.5.0", "1.6.0-alpha.1"})
	require.NoError(t, err)
	assert.Equal(t, "1.5.0", found.GetVersion())
}

// TestGreaterThanPatternWithoutDowngradeRejectsLowerPrereleases checks that the "prerelease"
// pattern never accepts a prerelease of the release in use, which is lower than it.
func TestGreaterThanPatternWithoutDowngradeRejectsLowerPrereleases(t *testing.T) {
	filter := Filter{Kind: SEMVERVERSIONKIND, Pattern: "prerelease"}
	pattern, err := filter.GreaterThanPatternWithoutDowngrade("1.2.3")
	require.NoError(t, err)

	bounded := Filter{Kind: SEMVERVERSIONKIND, Pattern: pattern}
	assert.False(t, bounded.Matches([]string{"1.2.2", "1.2.3-alpha", "1.2.3-rc.1"}))

	found, err := bounded.Search([]string{"1.2.3-alpha", "1.2.3"})
	require.NoError(t, err)
	assert.Equal(t, "1.2.3", found.GetVersion())

	// Without an age filter, the pattern is left as it was
	legacy, err := filter.GreaterThanPattern("1.2.3")
	require.NoError(t, err)
	assert.Equal(t, ">=1.2.3-0 <= 1.2.3", legacy)
}

// TestBoundPatternIgnoresBuildMetadataHyphens checks that a stable range with build metadata
// bounded at a prerelease still rejects prereleases.
func TestBoundPatternIgnoresBuildMetadataHyphens(t *testing.T) {
	filter := Filter{Kind: SEMVERVERSIONKIND, Pattern: BoundPattern("^1.2.0+2026-01", "1.5.0-beta.1")}

	assert.False(t, filter.Matches([]string{"1.4.0", "1.5.0-beta.1", "1.5.0-beta.2"}))

	found, err := filter.Search([]string{"1.4.0", "1.5.0-beta.2", "1.5.1", "1.6.0-alpha.1"})
	require.NoError(t, err)
	assert.Equal(t, "1.5.1", found.GetVersion())
}

func TestUnboundedPatternKeepsPrereleases(t *testing.T) {
	filter := Filter{Kind: SEMVERVERSIONKIND, Pattern: UnboundedPattern("2.0.0-rc.2")}

	// Without any newer version filtered out, the prerelease in use is kept
	found, err := filter.Search([]string{"1.9.0", "2.0.0-rc.1", "2.0.0-rc.2"})
	require.NoError(t, err)
	assert.Equal(t, "2.0.0-rc.2", found.GetVersion())
}

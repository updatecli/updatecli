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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()

			assert.Equal(t, tt.wantErr, err)
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
			name: "Minor semver pattern",
			filter: Filter{
				Kind:    SEMVERVERSIONKIND,
				Pattern: "minor",
			},
			version: "3.0", want: "3.x",
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
			name: "Patch semver pattern",
			filter: Filter{
				Kind:    SEMVERVERSIONKIND,
				Pattern: "patch",
			},
			version: "3.0", want: "3.0.x",
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

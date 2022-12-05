package version

import (
	"errors"
	"fmt"
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
			wantErr:  fmt.Errorf(`no version found matching pattern "^updatecli-4.(\\d*)$"`),
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
			wantErr: errors.New(`unsupported version kind "noExist"`),
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

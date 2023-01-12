package cargopackage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestPackageDir(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		packageDir  string
		wantErr     bool
	}{
		{
			name:        "Alphanumeric",
			packageName: "random",
			packageDir:  "ra/nd",
			wantErr:     false,
		},
		{
			name:        "4 Characters",
			packageName: "rand",
			packageDir:  "ra/nd",
			wantErr:     false,
		},
		{
			name:        "Special Char",
			packageName: "b-crypt65",
			packageDir:  "b-/cr",
			wantErr:     false,
		},
		{
			name:        "One Character",
			packageName: "a",
			packageDir:  "1",
			wantErr:     false,
		},
		{
			name:        "Two Characters",
			packageName: "az",
			packageDir:  "2",
			wantErr:     false,
		},
		{
			name:        "Three Characters",
			packageName: "zac",
			packageDir:  "3/z",
			wantErr:     false,
		},
		{
			name:        "Empty package",
			packageName: "",
			packageDir:  "",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := getPackageFileDir(tt.packageName)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, tt.packageDir, got)
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name              string
		spec              Spec
		wantSpec          Spec
		wantVersionFilter version.Filter
		wantErr           bool
	}{
		{
			name: "Normal case with default index",
			spec: Spec{
				Package: "rand",
			},
			wantSpec: Spec{
				Package: "rand",
			},
			wantVersionFilter: version.Filter{
				Kind:    "latest",
				Pattern: "latest",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := New(tt.spec, false)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantSpec, got.spec)
			assert.Equal(t, tt.wantVersionFilter, got.versionFilter)
		})
	}
}

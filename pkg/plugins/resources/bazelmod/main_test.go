package bazelmod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		spec    Spec
		wantErr bool
	}{
		{
			name: "Nominal case",
			spec: Spec{
				File:   "MODULE.bazel",
				Module: "rules_go",
			},
			wantErr: false,
		},
		{
			name: "Missing file",
			spec: Spec{
				Module: "rules_go",
			},
			wantErr: true,
		},
		{
			name: "Missing module",
			spec: Spec{
				File: "MODULE.bazel",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tt.spec.File, got.spec.File)
			assert.Equal(t, tt.spec.Module, got.spec.Module)
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		spec    Spec
		wantErr bool
	}{
		{
			name: "Valid spec",
			spec: Spec{
				File:   "MODULE.bazel",
				Module: "rules_go",
			},
			wantErr: false,
		},
		{
			name: "Missing file",
			spec: Spec{
				Module: "rules_go",
			},
			wantErr: true,
		},
		{
			name: "Missing module",
			spec: Spec{
				File: "MODULE.bazel",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := New(tt.spec)
			if tt.wantErr {
				// If New fails, Validate will also fail
				if err == nil {
					err = b.Validate()
					assert.Error(t, err)
				} else {
					assert.Error(t, err)
				}
				return
			}

			require.NoError(t, err)
			err = b.Validate()
			require.NoError(t, err)
		})
	}
}

func TestReportConfig(t *testing.T) {
	spec := Spec{
		File:   "MODULE.bazel",
		Module: "rules_go",
	}

	b, err := New(spec)
	require.NoError(t, err)

	reportConfig := b.ReportConfig()
	assert.NotNil(t, reportConfig)

	reportedSpec, ok := reportConfig.(Spec)
	require.True(t, ok)
	assert.Equal(t, spec.File, reportedSpec.File)
	assert.Equal(t, spec.Module, reportedSpec.Module)
}

func TestChangelog(t *testing.T) {
	spec := Spec{
		File:   "MODULE.bazel",
		Module: "rules_go",
	}

	b, err := New(spec)
	require.NoError(t, err)

	changelog := b.Changelog("0.42.0", "0.43.0")
	assert.Nil(t, changelog)
}

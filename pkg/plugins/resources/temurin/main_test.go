package temurin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		spec     Spec
		wantSpec Spec
		wantErr  string
	}{
		{
			name: "Normal case with default values",
			spec: Spec{},
			wantSpec: Spec{
				Architecture:    "x64",
				OperatingSystem: "linux",
				ImageType:       "jdk",
				ReleaseType:     "ga",
				Result:          "version",
				Project:         "jdk",
			},
		},
		{
			name: "Failing case with invalid Temurin architecture",
			spec: Spec{
				Architecture: "apple-silicon",
			},
			wantErr: "[temurin] Specified architecture \"apple-silicon\" is not a valid Temurin architecture ([x64 x32 ppc64 ppc64le s390x aarch64 arm sparcv9 riscv64])",
		},
		{
			name: "Failing case with invalid Temurin os",
			spec: Spec{
				OperatingSystem: "archlinux",
			},
			wantErr: "[temurin] Specified operating system ('os') \"archlinux\" is not a valid Temurin architecture ([linux windows mac solaris aix alpine-linux])",
		},
		{
			name: "Failing case with invalid Temurin release type",
			spec: Spec{
				OperatingSystem: "foo",
			},
			wantErr: "[temurin] Specified operating system ('os') \"foo\" is not a valid Temurin architecture ([linux windows mac solaris aix alpine-linux])",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := New(tt.spec)
			if tt.wantErr != "" {
				require.Error(t, gotErr)
				assert.Equal(t, tt.wantErr, gotErr.Error())
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantSpec, got.spec)
		})
	}
}

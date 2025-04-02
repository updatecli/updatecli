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
			wantErr: "[temurin] Specified architecture \"apple-silicon\" is not a valid Temurin architecture (check https://api.adoptium.net/q/swagger-ui/#/Types for valid list)",
		},
		{
			name: "Failing case with invalid Temurin os",
			spec: Spec{
				OperatingSystem: "archlinux",
			},
			wantErr: "[temurin] Specified operating system \"archlinux\" is not a valid Temurin architecture (check https://api.adoptium.net/q/swagger-ui/#/Types for valid list)",
		},
		{
			name: "Failing case with invalid Temurin release type",
			spec: Spec{
				OperatingSystem: "foo",
			},
			wantErr: "[temurin] Specified operating system \"foo\" is not a valid Temurin architecture (check https://api.adoptium.net/q/swagger-ui/#/Types for valid list)",
		},
		{
			name: "Failing case with invalid platform type (1)",
			spec: Spec{
				Platforms: []string{"linux/x64", "aarch64/windows"},
			},
			wantErr: "[temurin] Specified platform \"aarch64/windows\" is not a valid Temurin platform: \"aarch64\" is not a valid Operating System (check https://api.adoptium.net/q/swagger-ui/#/Types for valid list)",
		},
		{
			name: "Failing case with invalid platform type (2)",
			spec: Spec{
				Platforms: []string{"linux"},
			},
			wantErr: "[temurin] Specified platform \"linux\" is not a valid Temurin platform: it misses a CPU architecture",
		},
		{
			name: "Failing case with invalid platform type (3)",
			spec: Spec{
				Platforms: []string{"linux/x64/aarch64"},
			},
			wantErr: "[temurin] Specified platform \"linux/x64/aarch64\" is not a valid Temurin platform: too much items specified",
		},
		{
			name: "Failing case with invalid platform type (4)",
			spec: Spec{
				Platforms: []string{"linux/foo"},
			},
			wantErr: "[temurin] Specified platform \"linux/foo\" is not a valid Temurin platform: \"foo\" is not a valid Architecture (check https://api.adoptium.net/q/swagger-ui/#/Types for valid list)",
		},
		{
			name: "Failing case with both platforms and OS specified",
			spec: Spec{
				OperatingSystem: "linux",
				Platforms:       []string{"linux/x64"},
			},
			wantErr: "[temurin] `spec.platform` and `spec.operatingsystem` are mutually exclusive",
		},
		{
			name: "Failing case with both platforms and Architecture specified",
			spec: Spec{
				Architecture: "x64",
				Platforms:    []string{"linux/x64"},
			},
			wantErr: "[temurin] `spec.platform` and `spec.architecture` are mutually exclusive",
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

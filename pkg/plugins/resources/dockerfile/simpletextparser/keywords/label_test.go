package keywords

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabel_ReplaceLine(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		originalLine string
		matcher      string
		want         string
	}{
		{
			name:         "Match and change",
			source:       "0.14.6",
			originalLine: "LABEL org.opencontainers.image.version=0.13.6",
			matcher:      "org.opencontainers.image.version",
			want:         "LABEL org.opencontainers.image.version=0.14.6",
		},
		{
			name:         "Match and change (lower case instruction)",
			source:       "0.14.6",
			originalLine: "label org.opencontainers.image.version=0.13.6",
			matcher:      "org.opencontainers.image.version",
			want:         "label org.opencontainers.image.version=0.14.6",
		},
		{
			name:         "Match default value and change",
			source:       "0.14.6",
			originalLine: "LABEL org.opencontainers.image.version",
			matcher:      "org.opencontainers.image.version",
			want:         "LABEL org.opencontainers.image.version=0.14.6",
		},
		{
			name:         "Match empty value and change",
			source:       "0.14.6",
			originalLine: "LABEL org.opencontainers.image.version=",
			matcher:      "org.opencontainers.image.version",
			want:         "LABEL org.opencontainers.image.version=0.14.6",
		},
		{
			name:         "Match but no change",
			source:       "0.14.6",
			originalLine: "LABEL org.opencontainers.image.version=0.14.6",
			matcher:      "org.opencontainers.image.version",
			want:         "LABEL org.opencontainers.image.version=0.14.6",
		},
		{
			name:         "No Match at all",
			source:       "0.14.6",
			originalLine: "LABEL org.opencontainers.image.url=https://github.com/updatecli/updatecli",
			matcher:      "org.opencontainers.image.version",
			want:         "LABEL org.opencontainers.image.url=https://github.com/updatecli/updatecli",
		},
		{
			name:         "No Match for key but value, no change",
			source:       "0.14.6",
			originalLine: "LABEL org.opencontainers.image.base.image_version=TERRAFORM_VERSION",
			matcher:      "TERRAFORM_VERSION",
			want:         "LABEL org.opencontainers.image.base.image_version=TERRAFORM_VERSION",
		},
		{
			name:         "No Match for upper case key",
			source:       "0.14.6",
			originalLine: "LABEL org.opencontainers.image.version=0.13.6",
			matcher:      "ORG.OPENCONTAINERS.IMAGE.VERSION",
			want:         "LABEL org.opencontainers.image.version=0.13.6",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Label{}

			got := a.ReplaceLine(tt.source, tt.originalLine, tt.matcher)
			assert.Equal(t, tt.want, got)

		})
	}
}

func TestLabel_IsLineMatching(t *testing.T) {
	tests := []struct {
		name         string
		originalLine string
		matcher      string
		want         bool
	}{
		{
			name:         "Match",
			originalLine: "LABEL org.opencontainers.image.version=0.13.6",
			matcher:      "org.opencontainers.image.version",
			want:         true,
		},
		{
			name:         "Match (lower case instruction)",
			originalLine: "label org.opencontainers.image.version=0.13.6",
			matcher:      "org.opencontainers.image.version",
			want:         true,
		},
		{
			name:         "Match empty value",
			originalLine: "LABEL org.opencontainers.image.version=",
			matcher:      "org.opencontainers.image.version",
			want:         true,
		},
		{
			name:         "No Match at all",
			originalLine: "LABEL org.opencontainers.image.url=https://github.com/updatecli/updatecli",
			matcher:      "org.opencontainers.image.version",
			want:         false,
		},
		{
			name:         "No Match for key but value",
			originalLine: "LABEL org.opencontainers.image.base.image_version=TERRAFORM_VERSION",
			matcher:      "TERRAFORM_VERSION",
			want:         false,
		},
		{
			name:         "No Match for lower case key",
			originalLine: "LABEL org.opencontainers.image.version=0.13.6",
			matcher:      "ORG.OPENCONTAINERS.IMAGE.VERSION",
			want:         false,
		},
		{
			name:         "Empty line",
			originalLine: "",
			matcher:      "org.opencontainers.image.version",
			want:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Label{}

			got := a.IsLineMatching(tt.originalLine, tt.matcher)
			assert.Equal(t, tt.want, got)

		})
	}
}

func TestLabel_GetValue(t *testing.T) {
	tests := []struct {
		name         string
		originalLine string
		matcher      string
		want         string
		wantErr      bool
	}{
		{
			name:         "Match",
			originalLine: "LABEL org.opencontainers.image.version=0.13.6",
			matcher:      "org.opencontainers.image.version",
			want:         "0.13.6",
			wantErr:      false,
		},
		{
			name:         "Match (lower case instruction)",
			originalLine: "label org.opencontainers.image.version=0.13.6",
			matcher:      "org.opencontainers.image.version",
			want:         "0.13.6",
			wantErr:      false,
		},
		{
			name:         "Match empty value",
			originalLine: "LABEL org.opencontainers.image.version=",
			matcher:      "org.opencontainers.image.version",
			want:         "",
			wantErr:      false,
		},
		{
			name:         "No Match at all",
			originalLine: "LABEL org.opencontainers.image.url=https://github.com/updatecli/updatecli",
			matcher:      "org.opencontainers.image.version",
			want:         "",
			wantErr:      true,
		},
		{
			name:         "No Match for key but value",
			originalLine: "LABEL org.opencontainers.image.base.image_version=TERRAFORM_VERSION",
			matcher:      "TERRAFORM_VERSION",
			want:         "",
			wantErr:      true,
		},
		{
			name:         "No Match for lower case key",
			originalLine: "LABEL org.opencontainers.image.version=0.13.6",
			matcher:      "ORG.OPENCONTAINERS.IMAGE.VERSION",
			want:         "",
			wantErr:      true,
		},
		{
			name:         "Empty line",
			originalLine: "",
			matcher:      "org.opencontainers.image.version",
			want:         "",
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Label{}

			got, gotErr := a.GetValue(tt.originalLine, tt.matcher)
			if tt.wantErr {
				assert.Error(t, gotErr)
			} else {
				assert.Equal(t, tt.want, got)
			}

		})
	}
}

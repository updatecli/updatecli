package keywords

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArg_ReplaceLine(t *testing.T) {
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
			originalLine: "ARG TERRAFORM_VERSION=0.13.6",
			matcher:      "TERRAFORM_VERSION",
			want:         "ARG TERRAFORM_VERSION=0.14.6",
		},
		{
			name:         "Match and change (lower case instruction)",
			source:       "0.14.6",
			originalLine: "arg TERRAFORM_VERSION=0.13.6",
			matcher:      "TERRAFORM_VERSION",
			want:         "arg TERRAFORM_VERSION=0.14.6",
		},
		{
			name:         "Match default value and change",
			source:       "0.14.6",
			originalLine: "ARG TERRAFORM_VERSION",
			matcher:      "TERRAFORM_VERSION",
			want:         "ARG TERRAFORM_VERSION=0.14.6",
		},
		{
			name:         "Match empty value and change",
			source:       "0.14.6",
			originalLine: "ARG TERRAFORM_VERSION=",
			matcher:      "TERRAFORM_VERSION",
			want:         "ARG TERRAFORM_VERSION=0.14.6",
		},
		{
			name:         "Match but no change",
			source:       "0.14.6",
			originalLine: "ARG TERRAFORM_VERSION=0.14.6",
			matcher:      "TERRAFORM_VERSION",
			want:         "ARG TERRAFORM_VERSION=0.14.6",
		},
		{
			name:         "No Match at all",
			source:       "0.14.6",
			originalLine: "ARG GOLANG_VERSION=1.15.7",
			matcher:      "TERRAFORM_VERSION",
			want:         "ARG GOLANG_VERSION=1.15.7",
		},
		{
			name:         "No Match for key but value, no change",
			source:       "0.14.6",
			originalLine: "ARG GOLANG_VERSION=TERRAFORM_VERSION",
			matcher:      "TERRAFORM_VERSION",
			want:         "ARG GOLANG_VERSION=TERRAFORM_VERSION",
		},
		{
			name:         "No Match for lower case key",
			source:       "0.14.6",
			originalLine: "ARG terraform_version=0.13.6",
			matcher:      "TERRAFORM_VERSION",
			want:         "ARG terraform_version=0.13.6",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := SimpleKeyword{Keyword: "arg"}

			got := a.ReplaceLine(tt.source, tt.originalLine, tt.matcher)
			assert.Equal(t, tt.want, got)

		})
	}
}

func TestArg_IsLineMatching(t *testing.T) {
	tests := []struct {
		name         string
		originalLine string
		matcher      string
		want         bool
	}{
		{
			name:         "Match",
			originalLine: "ARG TERRAFORM_VERSION=0.13.6",
			matcher:      "TERRAFORM_VERSION",
			want:         true,
		},
		{
			name:         "Match (lower case instruction)",
			originalLine: "arg TERRAFORM_VERSION=0.13.6",
			matcher:      "TERRAFORM_VERSION",
			want:         true,
		},
		{
			name:         "Match default value",
			originalLine: "ARG TERRAFORM_VERSION",
			matcher:      "TERRAFORM_VERSION",
			want:         true,
		},
		{
			name:         "Match empty value",
			originalLine: "ARG TERRAFORM_VERSION=",
			matcher:      "TERRAFORM_VERSION",
			want:         true,
		},
		{
			name:         "No Match at all",
			originalLine: "ARG GOLANG_VERSION=1.15.7",
			matcher:      "TERRAFORM_VERSION",
			want:         false,
		},
		{
			name:         "No Match for key but value",
			originalLine: "ARG GOLANG_VERSION=TERRAFORM_VERSION",
			matcher:      "TERRAFORM_VERSION",
			want:         false,
		},
		{
			name:         "No Match for lower case key",
			originalLine: "ARG terraform_version=0.13.6",
			matcher:      "TERRAFORM_VERSION",
			want:         false,
		},
		{
			name:         "Empty line",
			originalLine: "",
			matcher:      "TERRAFORM_VERSION",
			want:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := SimpleKeyword{Keyword: "arg"}

			got := a.IsLineMatching(tt.originalLine, tt.matcher)
			assert.Equal(t, tt.want, got)

		})
	}
}

func TestArg_GetValue(t *testing.T) {
	tests := []struct {
		name         string
		originalLine string
		matcher      string
		want         string
		wantErr      bool
	}{
		{
			name:         "Match",
			originalLine: "ARG TERRAFORM_VERSION=0.13.6",
			matcher:      "TERRAFORM_VERSION",
			want:         "0.13.6",
		},
		{
			name:         "Match (lower case instruction)",
			originalLine: "arg TERRAFORM_VERSION=0.13.6",
			matcher:      "TERRAFORM_VERSION",
			want:         "0.13.6",
		},
		{
			name:         "Match default value",
			originalLine: "ARG TERRAFORM_VERSION",
			matcher:      "TERRAFORM_VERSION",
			want:         "",
		},
		{
			name:         "Match empty value",
			originalLine: "ARG TERRAFORM_VERSION=",
			matcher:      "TERRAFORM_VERSION",
			want:         "",
		},
		{
			name:         "No Match at all",
			originalLine: "ARG GOLANG_VERSION=1.15.7",
			matcher:      "TERRAFORM_VERSION",
			want:         "",
			wantErr:      true,
		},
		{
			name:         "No Match for key but value",
			originalLine: "ARG GOLANG_VERSION=TERRAFORM_VERSION",
			matcher:      "TERRAFORM_VERSION",
			want:         "",
			wantErr:      true,
		},
		{
			name:         "No Match for lower case key",
			originalLine: "ARG terraform_version=0.13.6",
			matcher:      "TERRAFORM_VERSION",
			want:         "",
			wantErr:      true,
		},
		{
			name:         "Empty line",
			originalLine: "",
			matcher:      "TERRAFORM_VERSION",
			want:         "",
			wantErr:      true,
		},
		{
			name:         "Match and comment",
			originalLine: "ARG TERRAFORM_VERSION=0.13.6 # Pin tf version",
			matcher:      "TERRAFORM_VERSION",
			want:         "0.13.6",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := SimpleKeyword{Keyword: "arg"}

			got, gotErr := a.GetValue(tt.originalLine, tt.matcher)
			if tt.wantErr {
				assert.Error(t, gotErr)
			} else {
				assert.Equal(t, tt.want, got)
			}

		})
	}
}

func TestArg_GetTokens(t *testing.T) {
	tests := []struct {
		name         string
		originalLine string
		want         *SimpleTokens
		wantErr      bool
	}{
		{
			name:         "Default Case",
			originalLine: "ARG TERRAFORM_VERSION=0.13.6",
			want: &SimpleTokens{
				Keyword: "ARG",
				Name:    "TERRAFORM_VERSION",
				Value:   "0.13.6",
			},
		},
		{
			name:         "Lower case",
			originalLine: "arg TERRAFORM_VERSION=0.13.6",
			want: &SimpleTokens{
				Keyword: "arg",
				Name:    "TERRAFORM_VERSION",
				Value:   "0.13.6",
			},
		},
		{
			name:         "No Value",
			originalLine: "ARG TERRAFORM_VERSION",
			want: &SimpleTokens{
				Keyword: "ARG",
				Name:    "TERRAFORM_VERSION",
				Value:   "",
			},
		},
		{
			name:         "Comment",
			originalLine: "ARG TERRAFORM_VERSION=0.13.6 # Pin tf version",
			want: &SimpleTokens{
				Keyword: "ARG",
				Name:    "TERRAFORM_VERSION",
				Value:   "0.13.6",
				Comment: "# Pin tf version",
			},
		},
		{
			name:         "Match and bad comment",
			originalLine: "ARG TERRAFORM_VERSION=0.13.6 // Pin tf version",
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := SimpleKeyword{Keyword: "arg"}

			got, gotErr := a.GetTokens(tt.originalLine)
			if tt.wantErr {
				assert.Error(t, gotErr)
			} else {
				assert.Equal(t, tt.want, got)
			}

		})
	}
}

func TestEnv_ReplaceLine(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		originalLine string
		matcher      string
		want         string
	}{
		{
			name:         "Match and change with equals sign",
			source:       "{VERSION_TERRA}",
			originalLine: "ENV TERRAFORM_VERSION={VERSIONTERRA}",
			matcher:      "TERRAFORM_VERSION",
			want:         "ENV TERRAFORM_VERSION={VERSION_TERRA}",
		},
		{
			name:         "Match and change without equals sign",
			source:       "{VERSION_TERRA}",
			originalLine: "ENV TERRAFORM_VERSION {VERSIONTERRA}",
			matcher:      "TERRAFORM_VERSION",
			want:         "ENV TERRAFORM_VERSION={VERSION_TERRA}",
		},
		{
			name:         "Match and change (lower case instruction)",
			source:       "0.14.6",
			originalLine: "env TERRAFORM_VERSION=0.13.6",
			matcher:      "TERRAFORM_VERSION",
			want:         "env TERRAFORM_VERSION=0.14.6",
		},
		{
			name:         "Match default value and change",
			source:       "0.14.6",
			originalLine: "ENV TERRAFORM_VERSION",
			matcher:      "TERRAFORM_VERSION",
			want:         "ENV TERRAFORM_VERSION=0.14.6",
		},
		{
			name:         "Match empty value and change",
			source:       "0.14.6",
			originalLine: "ENV TERRAFORM_VERSION=",
			matcher:      "TERRAFORM_VERSION",
			want:         "ENV TERRAFORM_VERSION=0.14.6",
		},
		{
			name:         "Match but no change",
			source:       "0.14.6",
			originalLine: "ENV TERRAFORM_VERSION=0.14.6",
			matcher:      "TERRAFORM_VERSION",
			want:         "ENV TERRAFORM_VERSION=0.14.6",
		},
		{
			name:         "No Match at all",
			source:       "0.14.6",
			originalLine: "ENV GOLANG_VERSION=1.15.7",
			matcher:      "TERRAFORM_VERSION",
			want:         "ENV GOLANG_VERSION=1.15.7",
		},
		{
			name:         "No Match for key but value, no change",
			source:       "0.14.6",
			originalLine: "ENV GOLANG_VERSION=TERRAFORM_VERSION",
			matcher:      "TERRAFORM_VERSION",
			want:         "ENV GOLANG_VERSION=TERRAFORM_VERSION",
		},
		{
			name:         "No Match for lower case key",
			source:       "0.14.6",
			originalLine: "ENV terraform_version=0.13.6",
			matcher:      "TERRAFORM_VERSION",
			want:         "ENV terraform_version=0.13.6",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := SimpleKeyword{Keyword: "env"}

			got := e.ReplaceLine(tt.source, tt.originalLine, tt.matcher)
			assert.Equal(t, tt.want, got)

		})
	}
}

func TestEnv_IsLineMatching(t *testing.T) {
	tests := []struct {
		name         string
		originalLine string
		matcher      string
		want         bool
	}{
		{
			name:         "Match",
			originalLine: "ENV TERRAFORM_VERSION=0.13.6",
			matcher:      "TERRAFORM_VERSION",
			want:         true,
		},
		{
			name:         "Match (lower case instruction)",
			originalLine: "env TERRAFORM_VERSION=0.13.6",
			matcher:      "TERRAFORM_VERSION",
			want:         true,
		},
		{
			name:         "Match default value",
			originalLine: "ENV TERRAFORM_VERSION",
			matcher:      "TERRAFORM_VERSION",
			want:         true,
		},
		{
			name:         "Match empty value",
			originalLine: "ENV TERRAFORM_VERSION=",
			matcher:      "TERRAFORM_VERSION",
			want:         true,
		},
		{
			name:         "No Match at all",
			originalLine: "ENV GOLANG_VERSION=1.15.7",
			matcher:      "TERRAFORM_VERSION",
			want:         false,
		},
		{
			name:         "No Match for key but value",
			originalLine: "ENV GOLANG_VERSION=TERRAFORM_VERSION",
			matcher:      "TERRAFORM_VERSION",
			want:         false,
		},
		{
			name:         "No Match for lower case key",
			originalLine: "ENV terraform_version=0.13.6",
			matcher:      "TERRAFORM_VERSION",
			want:         false,
		},
		{
			name:         "Empty line",
			originalLine: "",
			matcher:      "TERRAFORM_VERSION",
			want:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := SimpleKeyword{Keyword: "env"}

			got := e.IsLineMatching(tt.originalLine, tt.matcher)
			assert.Equal(t, tt.want, got)

		})
	}
}

func TestEnv_GetValue(t *testing.T) {
	tests := []struct {
		name         string
		originalLine string
		matcher      string
		want         string
		wantErr      bool
	}{
		{
			name:         "Match",
			originalLine: "ENV TERRAFORM_VERSION=0.13.6",
			matcher:      "TERRAFORM_VERSION",
			want:         "0.13.6",
			wantErr:      false,
		}, {
			name:         "Match Space (legacy)",
			originalLine: "ENV TERRAFORM_VERSION 0.13.6",
			matcher:      "TERRAFORM_VERSION",
			want:         "0.13.6",
			wantErr:      false,
		},

		{
			name:         "Match (lower case instruction)",
			originalLine: "env TERRAFORM_VERSION=0.13.6",
			matcher:      "TERRAFORM_VERSION",
			want:         "0.13.6",
			wantErr:      false,
		},
		{
			name:         "Match default value",
			originalLine: "ENV TERRAFORM_VERSION",
			matcher:      "TERRAFORM_VERSION",
			want:         "",
			wantErr:      false,
		},
		{
			name:         "Match empty value",
			originalLine: "ENV TERRAFORM_VERSION=",
			matcher:      "TERRAFORM_VERSION",
			want:         "",
			wantErr:      false,
		},
		{
			name:         "No Match at all",
			originalLine: "ENV GOLANG_VERSION=1.15.7",
			matcher:      "TERRAFORM_VERSION",
			want:         "",
			wantErr:      true,
		},
		{
			name:         "No Match for key but value",
			originalLine: "ENV GOLANG_VERSION=TERRAFORM_VERSION",
			matcher:      "TERRAFORM_VERSION",
			want:         "",
			wantErr:      true,
		},
		{
			name:         "No Match for lower case key",
			originalLine: "ENV terraform_version=0.13.6",
			matcher:      "TERRAFORM_VERSION",
			want:         "",
			wantErr:      true,
		},
		{
			name:         "Empty line",
			originalLine: "",
			matcher:      "TERRAFORM_VERSION",
			want:         "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := SimpleKeyword{Keyword: "env"}

			got, gotErr := a.GetValue(tt.originalLine, tt.matcher)
			if tt.wantErr {
				assert.Error(t, gotErr)
			} else {
				assert.Equal(t, tt.want, got)
			}

		})
	}
}

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
			a := SimpleKeyword{Keyword: "label"}

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
			a := SimpleKeyword{Keyword: "label"}

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
			a := SimpleKeyword{Keyword: "label"}

			got, gotErr := a.GetValue(tt.originalLine, tt.matcher)
			if tt.wantErr {
				assert.Error(t, gotErr)
			} else {
				assert.Equal(t, tt.want, got)
			}

		})
	}
}

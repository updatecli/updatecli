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
			a := Arg{}

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
			a := Arg{}

			got := a.IsLineMatching(tt.originalLine, tt.matcher)
			assert.Equal(t, tt.want, got)

		})
	}
}

package keywords

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFrom_ReplaceLine(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		originalLine string
		matcher      string
		want         string
	}{
		{
			name:         "Match and change",
			source:       "3.14",
			originalLine: "FROM alpine:3.12",
			matcher:      "alpine",
			want:         "FROM alpine:3.14",
		},
		{
			name:         "Match and change (lower case instruction)",
			source:       "3.14",
			originalLine: "from alpine:3.12",
			matcher:      "alpine",
			want:         "from alpine:3.14",
		},
		{
			name:         "Match and change with aliasing",
			source:       "3.14",
			originalLine: "FROM alpine:3.12 AS builder",
			matcher:      "alpine",
			want:         "FROM alpine:3.14 AS builder",
		},
		{
			name:         "Match and change with same name in alias",
			source:       "3.14",
			originalLine: "FROM alpine:3.12 AS alpine",
			matcher:      "alpine",
			want:         "FROM alpine:3.14 AS alpine",
		},
		{
			name:         "Match default tag and change",
			source:       "3.14",
			originalLine: "FROM alpine",
			matcher:      "alpine",
			want:         "FROM alpine:3.14",
		},
		{
			name:         "Match beginning value",
			source:       "3.14",
			originalLine: "FROM alpine-test",
			matcher:      "alpine",
			want:         "FROM alpine-test:3.14",
		},
		{
			name:         "Match but no change",
			source:       "3.12",
			originalLine: "FROM alpine:3.12",
			matcher:      "alpine",
			want:         "FROM alpine:3.12",
		},
		{
			name:         "No Match at all",
			source:       "3.13",
			originalLine: "FROM ubuntu:20.04",
			matcher:      "alpine",
			want:         "FROM ubuntu:20.04",
		},
		{
			name:         "No Match for key but value, no change",
			source:       "3.13",
			originalLine: "FROM ubuntu:alpine",
			matcher:      "alpine",
			want:         "FROM ubuntu:alpine",
		},
		{
			name:         "No Match for upper case key",
			source:       "3.14",
			originalLine: "FROM ALPINE:3.12",
			matcher:      "alpine",
			want:         "FROM ALPINE:3.12",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := From{}

			got := f.ReplaceLine(tt.source, tt.originalLine, tt.matcher)
			assert.Equal(t, tt.want, got)

		})
	}
}

func TestFrom_IsLineMatching(t *testing.T) {
	tests := []struct {
		name         string
		originalLine string
		matcher      string
		want         bool
	}{
		{
			name:         "Match",
			originalLine: "FROM alpine:3.12",
			matcher:      "alpine",
			want:         true,
		},
		{
			name:         "Match (lower case instruction)",
			originalLine: "from alpine:3.12",
			matcher:      "alpine",
			want:         true,
		},
		{
			name:         "Match with aliasing",
			originalLine: "FROM alpine:3.12 AS builder",
			matcher:      "alpine",
			want:         true,
		},
		{
			name:         "Match and change with same name in alias",
			originalLine: "FROM alpine:3.12 AS alpine",
			matcher:      "alpine",
			want:         true,
		},
		{
			name:         "Match default tag and change",
			originalLine: "FROM alpine",
			matcher:      "alpine",
			want:         true,
		},
		{
			name:         "Match beginning value",
			originalLine: "FROM alpine-test",
			matcher:      "alpine",
			want:         true,
		},
		{
			name:         "No Match at all",
			originalLine: "FROM ubuntu:20.04",
			matcher:      "alpine",
			want:         false,
		},
		{
			name:         "No Match for key but value",
			originalLine: "FROM ubuntu:alpine",
			matcher:      "alpine",
			want:         false,
		},
		{
			name:         "No Match for upper case key",
			originalLine: "FROM ALPINE:3.12",
			matcher:      "alpine",
			want:         false,
		},
		{
			name:         "No Match for alias named as matcher",
			originalLine: "FROM ubuntu:20.04 AS alpine",
			matcher:      "alpine",
			want:         false,
		},
		{
			name:         "Empty line",
			originalLine: "",
			matcher:      "alpine",
			want:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := From{}

			got := f.IsLineMatching(tt.originalLine, tt.matcher)
			assert.Equal(t, tt.want, got)

		})
	}
}

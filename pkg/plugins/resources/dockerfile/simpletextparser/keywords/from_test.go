package keywords

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFrom_GetToken(t *testing.T) {
	tests := []struct {
		name         string
		originalLine string
		want         fromToken
		wantErr      error
	}{
		{
			name:         "Empty line",
			originalLine: "",
			wantErr:      fmt.Errorf("Got an empty or malformed line"),
		},
		{
			name:         "Non FROM line",
			originalLine: "ARG TERRAFORM_VERSION",
			wantErr:      fmt.Errorf("Not a FROM line: \"ARG\""),
		},
		{
			name:         "Malformed FROM line",
			originalLine: "FROM",
			wantErr:      fmt.Errorf("Got an empty or malformed line"),
		},
		{
			name:         "No image in platformed line",
			originalLine: "FROM --platform=linux/amd64",
			wantErr:      fmt.Errorf("No image in line"),
		},
		{
			name:         "Malformed alias",
			originalLine: "FROM alpine AS",
			wantErr:      fmt.Errorf("Malformed FROM line, AS keyword but no value for it"),
		},
		{
			name:         "Malformed comment",
			originalLine: "FROM alpine malformed comment",
			wantErr:      fmt.Errorf("Remaining token in line that should be a comment but doesn't start with \"#\""),
		},
		{
			name:         "Correct comment",
			originalLine: "FROM alpine # correct comment",
			want:         fromToken{keyword: "FROM", image: "alpine", comment: "# correct comment"},
		},
		{
			name:         "Match Spec with platform no tag no digest with alias",
			originalLine: "FROM --platform=linux/amd64 alpine AS builder",
			want:         fromToken{keyword: "FROM", platform: "--platform=linux/amd64", image: "alpine", alias: "builder", aliasKw: "AS"},
		},
		{
			name:         "Match Spec with platform no tag no digest with alias lowercase",
			originalLine: "from --platform=linux/amd64 alpine as builder",
			want:         fromToken{keyword: "from", platform: "--platform=linux/amd64", image: "alpine", alias: "builder", aliasKw: "as"},
		},
		{
			name:         "Match Spec with platform no tag no digest no alias",
			originalLine: "FROM --platform=linux/amd64 alpine",
			want:         fromToken{keyword: "FROM", platform: "--platform=linux/amd64", image: "alpine"},
		},
		{
			name:         "Match Spec no platform no tag no digest with alias",
			originalLine: "FROM alpine AS builder",
			want:         fromToken{keyword: "FROM", image: "alpine", alias: "builder", aliasKw: "AS"},
		},
		{
			name:         "Match Spec no platform no tag no digest no alias",
			originalLine: "FROM alpine",
			want:         fromToken{keyword: "FROM", image: "alpine"},
		},
		{
			name:         "Match Spec with platform with tag no digest with alias",
			originalLine: "FROM --platform=linux/amd64 alpine:3.12 AS builder",
			want:         fromToken{keyword: "FROM", platform: "--platform=linux/amd64", image: "alpine", tag: "3.12", alias: "builder", aliasKw: "AS"},
		},
		{
			name:         "Match Spec with platform with tag no digest no alias",
			originalLine: "FROM --platform=linux/amd64 alpine:3.12",
			want:         fromToken{keyword: "FROM", platform: "--platform=linux/amd64", image: "alpine", tag: "3.12"},
		},
		{
			name:         "Match Spec no platform with tag no digest no alias",
			originalLine: "FROM alpine:3.12 AS builder",
			want:         fromToken{keyword: "FROM", image: "alpine", tag: "3.12", alias: "builder", aliasKw: "AS"},
		},
		{
			name:         "Match Spec no platform with tag no digest no alias",
			originalLine: "FROM alpine:3.12",
			want:         fromToken{keyword: "FROM", image: "alpine", tag: "3.12"},
		},
		{
			name:         "Match Spec with platform no tag with digest with alias",
			originalLine: "FROM --platform=linux/amd64 alpine@sha256:732 AS builder",
			want:         fromToken{keyword: "FROM", platform: "--platform=linux/amd64", image: "alpine", digest: "sha256:732", alias: "builder", aliasKw: "AS"},
		},
		{
			name:         "Match Spec with platform no tag with digest no alias",
			originalLine: "FROM --platform=linux/amd64 alpine@sha256:732",
			want:         fromToken{keyword: "FROM", platform: "--platform=linux/amd64", image: "alpine", digest: "sha256:732"},
		},
		{
			name:         "Match Spec no platform no tag with digest with alias",
			originalLine: "FROM alpine@sha256:732 AS builder",
			want:         fromToken{keyword: "FROM", image: "alpine", digest: "sha256:732", alias: "builder", aliasKw: "AS"},
		},
		{
			name:         "Match Spec no platform no tag with digest no alias",
			originalLine: "FROM alpine@sha256:732",
			want:         fromToken{keyword: "FROM", image: "alpine", digest: "sha256:732"},
		},
		{
			name:         "Match Spec with platform with tag with digest with alias",
			originalLine: "FROM --platform=linux/amd64 alpine:3.12@sha256:732 AS builder",
			want:         fromToken{keyword: "FROM", platform: "--platform=linux/amd64", image: "alpine", tag: "3.12", digest: "sha256:732", alias: "builder", aliasKw: "AS"},
		},
		{
			name:         "Match Spec with platform with tag with digest no alias",
			originalLine: "FROM --platform=linux/amd64 alpine:3.12@sha256:732",
			want:         fromToken{keyword: "FROM", platform: "--platform=linux/amd64", image: "alpine", tag: "3.12", digest: "sha256:732"},
		},
		{
			name:         "Match Spec no platform with tag with digest with alias",
			originalLine: "FROM alpine:3.12@sha256:732 AS builder",
			want:         fromToken{keyword: "FROM", image: "alpine", tag: "3.12", digest: "sha256:732", alias: "builder", aliasKw: "AS"},
		},
		{
			name:         "Match Spec no platform with tag with digest no alias",
			originalLine: "FROM alpine:3.12@sha256:732",
			want:         fromToken{keyword: "FROM", image: "alpine", tag: "3.12", digest: "sha256:732"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := From{}

			got, gotErr := f.parseTokens(tt.originalLine)
			if tt.wantErr != nil {
				assert.Error(t, gotErr)
				assert.Equal(t, tt.wantErr, gotErr)
				return
			}
			assert.Equal(t, tt.want, got)

		})
	}
}

func TestFrom_ReplaceLine(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		originalLine string
		matcher      string
		alias        bool
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
			name:         "Match and change alias with aliasing",
			source:       "newBuilder",
			originalLine: "FROM alpine:3.12 AS builder",
			matcher:      "builder",
			want:         "FROM alpine:3.12 AS newBuilder",
			alias:        true,
		},
		{
			name:         "No change alias without aliasing",
			source:       "newBuilder",
			originalLine: "FROM alpine:3.12",
			matcher:      "builder",
			want:         "FROM alpine:3.12",
			alias:        true,
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
		{
			name:         "No Match for upper case key and platform",
			source:       "3.14",
			originalLine: "FROM --platform=linux/amd64 ALPINE:3.12",
			matcher:      "alpine",
			want:         "FROM --platform=linux/amd64 ALPINE:3.12",
		},
		{
			name:         "Comment",
			source:       "3.14",
			originalLine: "FROM --platform=linux/amd64 alpine:3.12 # Base amd64 image",
			matcher:      "alpine",
			want:         "FROM --platform=linux/amd64 alpine:3.14 # Base amd64 image",
		},
		{
			name:         "Match Spec with platform no tag no digest with alias",
			source:       "3.14",
			originalLine: "FROM --platform=linux/amd64 alpine AS builder",
			matcher:      "alpine",
			want:         "FROM --platform=linux/amd64 alpine:3.14 AS builder",
		},
		{
			name:         "Match Spec with platform no tag no digest no alias",
			source:       "3.14",
			originalLine: "FROM --platform=linux/amd64 alpine",
			matcher:      "alpine",
			want:         "FROM --platform=linux/amd64 alpine:3.14",
		},
		{
			name:         "Match Spec no platform no tag no digest with alias",
			source:       "3.14",
			originalLine: "FROM alpine AS builder",
			matcher:      "alpine",
			want:         "FROM alpine:3.14 AS builder",
		},
		{
			name:         "Match Spec no platform no tag no digest no alias",
			source:       "3.14",
			originalLine: "FROM alpine",
			matcher:      "alpine",
			want:         "FROM alpine:3.14",
		},
		{
			name:         "Match Spec with platform with tag no digest with alias",
			source:       "3.14",
			originalLine: "FROM --platform=linux/amd64 alpine:3.12 AS builder",
			matcher:      "alpine",
			want:         "FROM --platform=linux/amd64 alpine:3.14 AS builder",
		},
		{
			name:         "Match Spec with platform with tag no digest no alias",
			source:       "3.14",
			originalLine: "FROM --platform=linux/amd64 alpine:3.12",
			matcher:      "alpine",
			want:         "FROM --platform=linux/amd64 alpine:3.14",
		},
		{
			name:         "Match Spec no platform with tag no digest no alias",
			source:       "3.14",
			originalLine: "FROM alpine:3.12 AS builder",
			matcher:      "alpine",
			want:         "FROM alpine:3.14 AS builder",
		},
		{
			name:         "Match Spec no platform with tag no digest no alias",
			source:       "3.14",
			originalLine: "FROM alpine:3.12",
			matcher:      "alpine",
			want:         "FROM alpine:3.14",
		},
		{
			name:         "Match Spec with platform no tag with digest with alias",
			source:       "@sha256:734",
			originalLine: "FROM --platform=linux/amd64 alpine@sha256:732 AS builder",
			matcher:      "alpine",
			want:         "FROM --platform=linux/amd64 alpine@sha256:734 AS builder",
		},
		{
			name:         "Match Spec with platform with tag with digest with alias",
			source:       "3.14@sha256:734",
			originalLine: "FROM --platform=linux/amd64 alpine:3.12@sha256:732 AS builder",
			matcher:      "alpine",
			want:         "FROM --platform=linux/amd64 alpine:3.14@sha256:734 AS builder",
		},
		{
			name:         "Match Spec with platform no tag with digest no alias",
			source:       "@sha256:734",
			originalLine: "FROM --platform=linux/amd64 alpine@sha256:732",
			matcher:      "alpine",
			want:         "FROM --platform=linux/amd64 alpine@sha256:734",
		},
		{
			name:         "Match Spec no platform no tag with digest with alias",
			source:       "@sha256:734",
			originalLine: "FROM alpine@sha256:732 AS builder",
			matcher:      "alpine",
			want:         "FROM alpine@sha256:734 AS builder",
		},
		{
			name:         "Match Spec no platform no tag with digest no alias",
			source:       "@sha256:734",
			originalLine: "FROM alpine@sha256:732",
			matcher:      "alpine",
			want:         "FROM alpine@sha256:734",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := From{Alias: tt.alias}

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
		alias        bool
		want         bool
	}{
		{
			name:         "Match",
			originalLine: "FROM alpine:3.12",
			matcher:      "alpine",
			want:         true,
		},
		{
			name:         "Match on alias",
			originalLine: "FROM alpine:3.12 AS builder",
			matcher:      "builder",
			alias:        true,
			want:         true,
		},
		{
			name:         "Non Match on alias",
			originalLine: "FROM alpine:3.12 AS builder",
			matcher:      "alpine",
			alias:        true,
			want:         false,
		},
		{
			name:         "Non Match on empty alias",
			originalLine: "FROM alpine:3.12",
			matcher:      "alpine",
			alias:        true,
			want:         false,
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
			name:         "Match and change with platform",
			originalLine: "FROM --platform=linux/amd64 alpine:3.12",
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
			f := From{Alias: tt.alias}

			got := f.IsLineMatching(tt.originalLine, tt.matcher)
			assert.Equal(t, tt.want, got)

		})
	}
}

func Test_GetStageName(t *testing.T) {
	tests := []struct {
		name         string
		originalLine string
		want         string
		wantErr      bool
	}{
		{
			name:         "No stage name",
			originalLine: "FROM alpine:3.12",
			want:         "",
		},
		{
			name:         "Non From line",
			originalLine: "ENV alpine=3.12",
			wantErr:      true,
		},
		{
			name:         "Empty line",
			originalLine: "",
			wantErr:      true,
		},
		{
			name:         "Stage Name",
			originalLine: "FROM alpine:3.12 AS builder",
			want:         "builder",
		},
		{
			name:         "Stage Name lowercase",
			originalLine: "from alpine:3.12 as builder",
			want:         "builder",
		},

		{
			name:         "Stage name and comment",
			originalLine: "FROM alpine:3.12 AS alpine",
			want:         "alpine",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := From{}

			got, gotErr := f.GetStageName(tt.originalLine)
			if tt.wantErr {
				assert.Error(t, gotErr)
			} else {
				assert.Equal(t, tt.want, got)
			}

		})
	}
}

func TestFrom_GetValue(t *testing.T) {
	tests := []struct {
		name         string
		originalLine string
		matcher      string
		stage        string
		alias        bool
		want         string
		wantErr      bool
	}{

		{
			name:         "Match",
			originalLine: "FROM alpine:3.12 AS builder",
			want:         "alpine:3.12",
			wantErr:      false,
		},
		{
			name:         "Match on alias",
			originalLine: "FROM alpine:3.12 AS builder",
			want:         "builder",
			alias:        true,
			wantErr:      false,
		},
		{
			name:         "Error on alias without alias",
			originalLine: "FROM alpine:3.12",
			want:         "alpine",
			alias:        true,
			wantErr:      true,
		},
		{
			name:         "Lowercase Match",
			originalLine: "from alpine:3.12 as builder",
			want:         "alpine:3.12",
			wantErr:      false,
		},
		{
			name:         "Empty line",
			originalLine: "",
			matcher:      "alpine",
			want:         "",
			wantErr:      true,
		},
		{
			name:         "Special Characters",
			originalLine: "from eclipse-temurin:${JAVA_VERSION}-jdk-jammy AS jdk",
			matcher:      "eclipse-temurin",
			want:         "eclipse-temurin:${JAVA_VERSION}-jdk-jammy",
			wantErr:      false,
		},
		{
			name:         "Platform",
			originalLine: "FROM --platform=linux/amd64 eclipse-mosquito:2.0",
			matcher:      "eclipse-mosquito",
			want:         "eclipse-mosquito:2.0",
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := From{Alias: tt.alias}

			got, gotErr := a.GetValue(tt.originalLine, tt.matcher)
			if tt.wantErr {
				assert.Error(t, gotErr)
			} else {
				assert.Equal(t, tt.want, got)
			}

		})
	}
}

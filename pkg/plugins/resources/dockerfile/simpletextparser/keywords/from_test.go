package keywords

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFrom_GetTokens(t *testing.T) {
	tests := []struct {
		name         string
		originalLine string
		want         FromToken
		wantErr      error
	}{
		{
			name:         "Empty line",
			originalLine: "",
			wantErr:      fmt.Errorf("got an empty or malformed line"),
		},
		{
			name:         "Non FROM line",
			originalLine: "ARG TERRAFORM_VERSION",
			wantErr:      fmt.Errorf("not a FROM line: \"ARG\""),
		},
		{
			name:         "Malformed FROM line",
			originalLine: "FROM",
			wantErr:      fmt.Errorf("got an empty or malformed line"),
		},
		{
			name:         "No image in platformed line",
			originalLine: "FROM --platform=linux/amd64",
			wantErr:      fmt.Errorf("no image in line"),
		},
		{
			name:         "Malformed alias",
			originalLine: "FROM alpine AS",
			wantErr:      fmt.Errorf("malformed FROM line, AS keyword but no value for it"),
		},
		{
			name:         "Malformed comment",
			originalLine: "FROM alpine malformed comment",
			wantErr:      fmt.Errorf("remaining token in line that should be a comment but doesn't start with \"#\""),
		},
		{
			name:         "Correct comment",
			originalLine: "FROM alpine # correct comment",
			want:         FromToken{Keyword: "FROM", Image: "alpine", Tag: "latest", Comment: "# correct comment"},
		},
		{
			name:         "Match Spec with platform no tag no digest with alias",
			originalLine: "FROM --platform=linux/amd64 alpine AS builder",
			want:         FromToken{Keyword: "FROM", Platform: "linux/amd64", Image: "alpine", Tag: "latest", Alias: "builder", AliasKw: "AS"},
		},
		{
			name:         "Match Spec with platform no tag no digest with alias lowercase",
			originalLine: "from --platform=linux/amd64 alpine as builder",
			want:         FromToken{Keyword: "from", Platform: "linux/amd64", Image: "alpine", Tag: "latest", Alias: "builder", AliasKw: "as"},
		},
		{
			name:         "Match Spec with platform no tag no digest no alias",
			originalLine: "FROM --platform=linux/amd64 alpine",
			want:         FromToken{Keyword: "FROM", Platform: "linux/amd64", Image: "alpine", Tag: "latest"},
		},
		{
			name:         "Match Spec no platform no tag no digest with alias",
			originalLine: "FROM alpine AS builder",
			want:         FromToken{Keyword: "FROM", Image: "alpine", Tag: "latest", Alias: "builder", AliasKw: "AS"},
		},
		{
			name:         "Match Spec no platform no tag no digest no alias",
			originalLine: "FROM alpine",
			want:         FromToken{Keyword: "FROM", Image: "alpine", Tag: "latest"},
		},
		{
			name:         "Match Spec with platform with tag no digest with alias",
			originalLine: "FROM --platform=linux/amd64 alpine:3.12 AS builder",
			want:         FromToken{Keyword: "FROM", Platform: "linux/amd64", Image: "alpine", Tag: "3.12", Alias: "builder", AliasKw: "AS"},
		},
		{
			name:         "Match Spec with platform with tag no digest no alias",
			originalLine: "FROM --platform=linux/amd64 alpine:3.12",
			want:         FromToken{Keyword: "FROM", Platform: "linux/amd64", Image: "alpine", Tag: "3.12"},
		},
		{
			name:         "Match Spec no platform with tag no digest no alias",
			originalLine: "FROM alpine:3.12 AS builder",
			want:         FromToken{Keyword: "FROM", Image: "alpine", Tag: "3.12", Alias: "builder", AliasKw: "AS"},
		},
		{
			name:         "Match Spec no platform with tag no digest no alias",
			originalLine: "FROM alpine:3.12",
			want:         FromToken{Keyword: "FROM", Image: "alpine", Tag: "3.12"},
		},
		{
			name:         "Match Spec with platform no tag with digest with alias",
			originalLine: "FROM --platform=linux/amd64 alpine@sha256:732 AS builder",
			want:         FromToken{Keyword: "FROM", Platform: "linux/amd64", Image: "alpine", Digest: "sha256:732", Alias: "builder", AliasKw: "AS"},
		},
		{
			name:         "Match Spec with platform no tag with digest no alias",
			originalLine: "FROM --platform=linux/amd64 alpine@sha256:732",
			want:         FromToken{Keyword: "FROM", Platform: "linux/amd64", Image: "alpine", Digest: "sha256:732"},
		},
		{
			name:         "Match Spec no platform no tag with digest with alias",
			originalLine: "FROM alpine@sha256:732 AS builder",
			want:         FromToken{Keyword: "FROM", Image: "alpine", Digest: "sha256:732", Alias: "builder", AliasKw: "AS"},
		},
		{
			name:         "Match Spec no platform no tag with digest no alias",
			originalLine: "FROM alpine@sha256:732",
			want:         FromToken{Keyword: "FROM", Image: "alpine", Digest: "sha256:732"},
		},
		{
			name:         "Match Spec with platform with tag with digest with alias",
			originalLine: "FROM --platform=linux/amd64 alpine:3.12@sha256:732 AS builder",
			want:         FromToken{Keyword: "FROM", Platform: "linux/amd64", Image: "alpine", Tag: "3.12", Digest: "sha256:732", Alias: "builder", AliasKw: "AS"},
		},
		{
			name:         "Match Spec with platform with tag with digest no alias",
			originalLine: "FROM --platform=linux/amd64 alpine:3.12@sha256:732",
			want:         FromToken{Keyword: "FROM", Platform: "linux/amd64", Image: "alpine", Tag: "3.12", Digest: "sha256:732"},
		},
		{
			name:         "Match Spec no platform with tag with digest with alias",
			originalLine: "FROM alpine:3.12@sha256:732 AS builder",
			want:         FromToken{Keyword: "FROM", Image: "alpine", Tag: "3.12", Digest: "sha256:732", Alias: "builder", AliasKw: "AS"},
		},
		{
			name:         "Match Spec no platform with tag with digest no alias",
			originalLine: "FROM alpine:3.12@sha256:732",
			want:         FromToken{Keyword: "FROM", Image: "alpine", Tag: "3.12", Digest: "sha256:732"},
		},
		{
			name:         "Parameterized Platform",
			originalLine: "FROM --platform=${PLATFORM} alpine:3.12",
			want: FromToken{
				Keyword:  "FROM",
				Image:    "alpine",
				Tag:      "3.12",
				Platform: "${PLATFORM}",
				Args: map[string]*FromTokenArgs{
					"platform": {
						Name: "PLATFORM",
					},
				},
			},
		},
		{
			name:         "Parameterized image",
			originalLine: "FROM ${image}:latest",
			want: FromToken{
				Keyword: "FROM",
				Image:   "${image}",
				Tag:     "latest",
				Args: map[string]*FromTokenArgs{
					"image": {
						Name: "image",
					},
				},
			},
		},
		{
			name:         "Parameterized version",
			originalLine: "FROM alpine:3.${version}",
			want: FromToken{
				Keyword: "FROM",
				Image:   "alpine",
				Tag:     "3.${version}",
				Args: map[string]*FromTokenArgs{
					"tag": {
						Prefix: "3.",
						Name:   "version",
					},
				},
			},
		},
		{
			name:         "Parameterized digest",
			originalLine: "FROM alpine:3@${digest}",
			want: FromToken{
				Keyword: "FROM",
				Image:   "alpine",
				Tag:     "3",
				Digest:  "${digest}",
				Args: map[string]*FromTokenArgs{
					"digest": {
						Name: "digest",
					},
				},
			},
		},
		{
			name:         "Full Parameterized",
			originalLine: "FROM --platform=${system}/amd64 jenkins-${jenkins-type}:${jenkins-version}@${jenkins-digest}",
			want: FromToken{
				Keyword:  "FROM",
				Platform: "${system}/amd64",
				Image:    "jenkins-${jenkins-type}",
				Tag:      "${jenkins-version}",
				Digest:   "${jenkins-digest}",
				Args: map[string]*FromTokenArgs{
					"platform": {
						Name:   "system",
						Suffix: "/amd64",
					},
					"image": {
						Prefix: "jenkins-",
						Name:   "jenkins-type",
					},
					"tag": {
						Name: "jenkins-version",
					},
					"digest": {
						Name: "jenkins-digest",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := From{}

			got, gotErr := f.GetTokens(tt.originalLine)
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
			f := From{}

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
			a := From{}

			got, gotErr := a.GetValue(tt.originalLine, tt.matcher)
			if tt.wantErr {
				assert.Error(t, gotErr)
			} else {
				assert.Equal(t, tt.want, got)
			}

		})
	}
}

func TestFrom_ParseTokenArgs(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		want    *FromTokenArgs
		wantErr bool
	}{

		{
			name:  "Default Case",
			token: "2.235-lts",
			want:  nil,
		},
		{
			name:  "Only Name",
			token: "${alpine_version}",
			want: &FromTokenArgs{
				Prefix: "",
				Suffix: "",
				Name:   "alpine_version",
			},
		},
		{
			name:    "Too many variables",
			token:   "${jenkins_release_type}-${jenkins_version}",
			want:    nil,
			wantErr: true,
		},
		{
			name:  "Only prefix",
			token: "alpine-${alpine_version}",
			want: &FromTokenArgs{
				Prefix: "alpine-",
				Suffix: "",
				Name:   "alpine_version",
			},
		},
		{
			name:  "Only suffix",
			token: "${jenkins_release_type}-latest",
			want: &FromTokenArgs{
				Prefix: "",
				Suffix: "-latest",
				Name:   "jenkins_release_type",
			},
		},
		{
			name:  "Prefix and Suffix",
			token: "linux-${jenkins_release_type}-latest",
			want: &FromTokenArgs{
				Prefix: "linux-",
				Suffix: "-latest",
				Name:   "jenkins_release_type",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := parseArg(tt.token)
			if tt.wantErr {
				assert.Error(t, gotErr)
			} else {
				assert.Equal(t, tt.want, got)
			}

		})
	}
}

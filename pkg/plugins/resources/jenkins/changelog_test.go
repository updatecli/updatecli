package jenkins

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestJenkins_Changelog(t *testing.T) {
	tests := []struct {
		name string
		sut  Jenkins
		from string
		to   string
		want *result.Changelogs
	}{
		{
			name: "Normal case with stable changelog",
			from: "2.39.2",
			to:   "2.39.2",
			sut: Jenkins{
				spec: Spec{Release: STABLE},
			},
			want: &result.Changelogs{
				{
					Title: "2.39.2",
					Body:  "Jenkins changelog is available at: https://www.jenkins.io/changelog-stable/#v2.39.2\n",
					URL:   "https://www.jenkins.io/changelog-stable/#v2.39.2",
				},
			},
		},
		{
			name: "Normal case with weekly changelog",
			from: "2.200",
			to:   "2.200",
			sut: Jenkins{
				spec: Spec{Release: WEEKLY},
			},
			want: &result.Changelogs{
				{
					Title: "2.200",
					Body:  "Jenkins changelog is available at: https://www.jenkins.io/changelog/#v2.200\n",
					URL:   "https://www.jenkins.io/changelog/#v2.200",
				},
			},
		},
		{
			name: "Error case with unknown baseline",
			from: "2.39.2",
			to:   "2.39.2",
			sut: Jenkins{
				spec: Spec{Release: "FOO"},
			},
			want: nil,
		},
		{
			name: "Error case with empty input release version",
			from: "",
			to:   "",
			sut: Jenkins{
				spec: Spec{Release: STABLE},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.sut.Changelog(tt.from, tt.to)

			assert.Equal(t, tt.want, got)
		})
	}
}

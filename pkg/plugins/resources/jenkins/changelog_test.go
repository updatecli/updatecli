package jenkins

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			from: "2.32.2",
			to:   "2.32.2",
			sut: Jenkins{
				spec: Spec{Release: STABLE},
			},
			want: &result.Changelogs{
				{
					Title: "2.32.2",
					Body:  "Jenkins changelog is available at: https://www.jenkins.io/changelog-stable/#v2.32.2\n",
					URL:   "https://www.jenkins.io/changelog-stable/#v2.32.2",
				},
			},
		},
		{
			name: "Version range case with stable changelog",
			from: "2.32.1",
			to:   "2.32.2",
			sut: Jenkins{
				spec: Spec{Release: STABLE},
			},
			want: &result.Changelogs{
				{
					Title: "2.32.2",
					Body:  "Jenkins changelog is available at: https://www.jenkins.io/changelog-stable/#v2.32.2\n",
					URL:   "https://www.jenkins.io/changelog-stable/#v2.32.2",
				},
				{
					Title: "2.32.1",
					Body:  "Jenkins changelog is available at: https://www.jenkins.io/changelog-stable/#v2.32.1\n",
					URL:   "https://www.jenkins.io/changelog-stable/#v2.32.1",
				},
			},
		},
		{
			name: "Case with unknown baseline",
			from: "xxx",
			to:   "2.32.2",
			sut: Jenkins{
				spec: Spec{Release: STABLE},
			},
			want: &result.Changelogs{
				{
					Title: "2.32.2",
					Body:  "Jenkins changelog is available at: https://www.jenkins.io/changelog-stable/#v2.32.2\n",
					URL:   "https://www.jenkins.io/changelog-stable/#v2.32.2",
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
			name: "Error case with wong input release version",
			from: "xxx",
			to:   "yyy",
			sut: Jenkins{
				spec: Spec{Release: STABLE},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j, err := New(tt.sut.spec)
			require.NoError(t, err)
			got := j.Changelog(tt.from, tt.to)

			assert.Equal(t, tt.want, got)
		})
	}
}

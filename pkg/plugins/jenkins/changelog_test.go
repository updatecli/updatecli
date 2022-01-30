package jenkins

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJenkins_Changelog(t *testing.T) {
	tests := []struct {
		name string
		sut  Jenkins
		want string
	}{
		{
			name: "Normal case with stable changelog",
			sut: Jenkins{
				spec:         Spec{Release: STABLE},
				foundVersion: "2.319.2",
			},
			want: "Jenkins changelog is available at: https://www.jenkins.io/changelog-stable/#v2.319.2\n",
		},
		{
			name: "Normal case with weekly changelog",
			sut: Jenkins{
				spec:         Spec{Release: WEEKLY},
				foundVersion: "2.200",
			},
			want: "Jenkins changelog is available at: https://www.jenkins.io/changelog/#v2.200\n",
		},
		{
			name: "Error case with unknown baseline",
			sut: Jenkins{
				spec:         Spec{Release: "FOO"},
				foundVersion: "2.319.2",
			},
			want: "",
		},
		{
			name: "Error case with empty input release version",
			sut: Jenkins{
				spec:         Spec{Release: STABLE},
				foundVersion: "",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.sut.Changelog()

			assert.Equal(t, tt.want, got)
		})
	}
}

package jenkins

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/version"
)

func TestJenkins_Changelog(t *testing.T) {
	tests := []struct {
		name    string
		spec    Spec
		release version.Version
		want    string
		wantErr bool
	}{
		{
			name:    "Normal case with stable changelog",
			spec:    Spec{Release: STABLE},
			release: version.Version{ParsedVersion: "2.319.2"},
			want:    "Jenkins changelog is available at: https://www.jenkins.io/changelog-stable/#v2.319.2\n",
		},
		{
			name:    "Normal case with weekly changelog",
			spec:    Spec{Release: WEEKLY},
			release: version.Version{ParsedVersion: "2.200"},
			want:    "Jenkins changelog is available at: https://www.jenkins.io/changelog/#v2.200\n",
		},
		{
			name:    "Error case with unknown baseline",
			spec:    Spec{Release: "FOO"},
			release: version.Version{ParsedVersion: "2.319.2"},
			wantErr: true,
		},
		{
			name:    "Error case with empty input release version",
			spec:    Spec{Release: STABLE},
			release: version.Version{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &Jenkins{
				spec: tt.spec,
			}
			got, err := j.Changelog(tt.release)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

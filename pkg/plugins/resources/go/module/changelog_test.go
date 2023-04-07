package gomodule

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestChangelog(t *testing.T) {
	tests := []struct {
		name           string
		version        GoModule
		expectedResult string
	}{
		{
			name: "Test getting changelog from github",
			version: GoModule{
				Spec: Spec{
					Module: "github.com/updatecli/updatecli",
				},
				Version: version.Version{
					OriginalVersion: "v0.42.0",
				},
			},
			expectedResult: "Changelog retrieved from:\n\thttps://github.com/updatecli/updatecli/releases/tag/v0.42.0\n## Changes\r\n\r\n## 🚀 Features\r\n\r\n- [source][condition] Add Cargo Package support @loispostula (#1081)\r\n\r\n## 🐛 Bug Fixes\r\n\r\n- chore sourceID deprecated for sourceid @pilere (#1070)\r\n- Only set git http auth if both username && password are specified @olblak (#1079)\r\n\r\n## 🧰 Maintenance\r\n\r\n- chore(deps): Bump github.com/aws/aws-sdk-go from 1.44.156 to 1.44.179 @dependabot (#1085)\r\n- chore(deps): Bump github.com/containerd/containerd from 1.6.6 to 1.6.12 @dependabot (#1082)\r\n- chore(deps): Bump updatecli/updatecli-action from 2.16.2 to 2.17.0 @dependabot (#1080)\r\n- chore(deps): Bump golang.org/x/oauth2 from 0.3.0 to 0.4.0 @dependabot (#1072)\r\n- chore(deps): Bump golang.org/x/text from 0.5.0 to 0.6.0 @dependabot (#1074)\r\n- chore(deps): Bump github.com/go-git/go-git/v5 from 5.5.1 to 5.5.2 @dependabot (#1075)\r\n\r\n## Contributors\r\n\r\n@dependabot, @dependabot[bot], @loispostula, @olblak, @pilere, @updateclibot and @updateclibot[bot]\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedResult, tt.version.Changelog())
		})
	}
}

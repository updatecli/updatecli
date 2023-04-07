package language

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestChangelog(t *testing.T) {
	tests := []struct {
		name           string
		version        Language
		expectedResult string
	}{
		{
			name: "Test new minor version",
			version: Language{
				Version: version.Version{
					OriginalVersion: "1.20",
				},
			},
			expectedResult: "Golang changelog for version \"1.20\" is available on \"https://go.dev/doc/go1.20\"",
		},
		{
			name: "Test new patch version",
			version: Language{
				Version: version.Version{
					OriginalVersion: "1.20.1",
				},
			},
			expectedResult: "Golang changelog for version \"1.20.1\" is available on \"https://go.dev/doc/devel/release#go1.20.minor\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedResult, tt.version.Changelog())
		})
	}
}

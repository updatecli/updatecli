package cargo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMatchingRule(t *testing.T) {

	dataset := []struct {
		name           string
		rules          MatchingRules
		filePath       string
		repository     string
		crateName      string
		crateVersion   string
		rootDir        string
		expectedResult bool
	}{
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/cargo.toml",
				},
			},
			filePath:       "testdata/cargo.toml",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/cargo.toml",
				},
			},
			filePath:       "./website/testdata/cargo.toml",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/cargo.toml",
					Crates: map[string]string{
						"clap": "",
					},
				},
			},
			filePath:       "testdata/cargo.toml",
			crateName:      "clap",
			expectedResult: true,
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			gotResult := d.rules.isMatchingRules(
				d.rootDir,
				d.filePath,
				d.repository,
				d.crateName,
				d.crateVersion)

			assert.Equal(t, d.expectedResult, gotResult)
		})
	}
}

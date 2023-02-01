package dockercompose

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMatchingRules(t *testing.T) {
	testdata := []struct {
		name           string
		rules          MatchingRules
		rootDir        string
		filePath       string
		service        string
		image          string
		arch           string
		expectedResult bool
	}{
		{
			name: "Scenario 1 - matching 1 rule",
			rules: MatchingRules{
				MatchingRule{
					Path: "docker-compose.yaml",
				},
			},
			filePath:       "docker-compose.yaml",
			expectedResult: true,
		},
		{
			name: "Scenario 2 - matching all rule",
			rules: MatchingRules{
				MatchingRule{
					Path: "docker-compose.yaml",
					Services: []string{
						"mongodb",
					},
				},
			},
			filePath:       "docker-compose.yaml",
			service:        "mongodb",
			expectedResult: true,
		},
		{
			name: "Scenario 3 - not matching all rules",
			rules: MatchingRules{
				MatchingRule{
					Path: "docker-compose.2.yaml",
					Services: []string{
						"mongodb",
					},
				},
			},
			filePath:       "docker-compose.yaml",
			service:        "mongodb",
			expectedResult: false,
		},
		{
			name: "Scenario 4 - only matching image name",
			rules: MatchingRules{
				MatchingRule{
					Images: []string{
						"mongo",
					},
				},
			},
			filePath:       "docker-compose.yaml",
			service:        "mongodb",
			image:          "mongo:6",
			expectedResult: true,
		},
		{
			name: "Scenario 5 - matching image name and tag",
			rules: MatchingRules{
				MatchingRule{
					Images: []string{
						"mongo:6",
					},
				},
			},
			filePath:       "docker-compose.yaml",
			service:        "mongodb",
			image:          "mongo:6",
			expectedResult: true,
		},
		{
			name: "Scenario 6 - correct image but wrong tag",
			rules: MatchingRules{
				MatchingRule{
					Images: []string{
						"mongo:6",
					},
				},
			},
			filePath:       "docker-compose.yaml",
			service:        "mongodb",
			image:          "mongo:7",
			expectedResult: false,
		},
		{
			name: "Scenario 6 - correct image and arch",
			rules: MatchingRules{
				MatchingRule{
					Images: []string{
						"mongo:6",
					},
					Archs: []string{
						"amd64",
					},
				},
			},
			filePath:       "docker-compose.yaml",
			service:        "mongodb",
			image:          "mongo:6",
			arch:           "amd64",
			expectedResult: true,
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {
			gotResult := tt.rules.isMatchingRule(
				tt.rootDir,
				tt.filePath,
				tt.service,
				tt.image,
				tt.arch)

			assert.Equal(t, tt.expectedResult, gotResult)

		})
	}

}

package maven

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMatchingRule(t *testing.T) {

	dataset := []struct {
		name            string
		rules           MatchingRules
		filePath        string
		groupid         string
		artifactName    string
		artifactVersion string
		rootDir         string
		expectedResult  bool
	}{
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/pom.xml",
				},
			},
			filePath:       "testdata/pom.xml",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/pom.xml",
				},
			},
			filePath:       "./website/testdata/pom.xml",
			expectedResult: false,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/pom.xml",
					ArtifactIDs: map[string]string{
						"jenkins-war": "",
					},
				},
			},
			filePath:       "testdata/pom.xml",
			artifactName:   "jenkins-war",
			expectedResult: true,
		},
		{
			rules: MatchingRules{
				MatchingRule{
					Path: "testdata/pom.xml",
					GroupIDs: []string{
						"junit",
					},
				},
			},
			filePath:       "testdata/pom.xml",
			groupid:        "junit",
			expectedResult: true,
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			gotResult := d.rules.isMatchingRules(
				d.rootDir,
				d.filePath,
				d.groupid,
				d.artifactName,
				d.artifactVersion)

			assert.Equal(t, d.expectedResult, gotResult)
		})
	}
}

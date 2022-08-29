package maven

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsRepositoriesContainsMavenCentral(t *testing.T) {
	testData := []struct {
		Name           string
		Repositories   []string
		ExpectedResult bool
		WantErr        bool
	}{
		{
			Name: "1",
			Repositories: []string{
				"repo.jenkins-ci.org",
			},
			ExpectedResult: false,
		},
		{
			Name: "1",
			Repositories: []string{
				"repo.jenkins-ci.org",
				"example.com",
			},
			ExpectedResult: false,
		},
		{
			Name: "1",
			Repositories: []string{
				"repo.jenkins-ci.org",
				"repo1.maven.org",
			},
			ExpectedResult: true,
		},
	}

	for _, tt := range testData {
		t.Run(tt.Name, func(t *testing.T) {
			gotResult, gotError := isRepositoriesContainsMavenCentral(tt.Repositories)

			if tt.WantErr {
				require.Error(t, gotError)
			} else {
				require.NoError(t, gotError)
			}

			assert.Equal(t, tt.ExpectedResult, gotResult)

		})
	}

}

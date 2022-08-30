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
				"https://repo.jenkins-ci.org",
			},
			ExpectedResult: false,
		},
		{
			Name: "1",
			Repositories: []string{
				"https://repo.jenkins-ci.org",
				"https://example.com",
			},
			ExpectedResult: false,
		},
		{
			Name: "1",
			Repositories: []string{
				"https://repo.jenkins-ci.org",
				"https://repo1.maven.org",
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

func TestGetURLHostname(t *testing.T) {
	testData := []struct {
		Name           string
		Repository     string
		ExpectedResult string
		WantErr        bool
	}{
		{
			Name:           "test 1",
			Repository:     "https://username:password@example.com",
			ExpectedResult: "https://example.com",
		},
		{
			Name:           "test 1",
			Repository:     "https://username:password@example.com",
			ExpectedResult: "https://example.com",
		},
		{
			Name:           "test 1",
			Repository:     "https://username:password@example.com/registry",
			ExpectedResult: "https://example.com/registry",
		},
	}

	for _, tt := range testData {
		t.Run(tt.Name, func(t *testing.T) {
			gotResult, gotError := trimUsernamePasswordFromURL(tt.Repository)

			if tt.WantErr {
				require.Error(t, gotError)
			} else {
				require.NoError(t, gotError)
			}

			assert.Equal(t, tt.ExpectedResult, gotResult)
		})
	}

}

func TestJoinURLElem(t *testing.T) {
	testData := []struct {
		Name           string
		Repository     []string
		ExpectedResult string
		WantErr        bool
	}{
		{
			Name:           "test 0",
			ExpectedResult: "",
		},
		{
			Name: "test 1",
			Repository: []string{
				"username:password@example.com",
				"registry",
			},
			ExpectedResult: "https://username:password@example.com/registry",
		},
		{
			Name: "test 2",
			Repository: []string{
				"username:password@example.com/",
				"registry",
			},
			ExpectedResult: "https://username:password@example.com/registry",
		},
		{
			Name: "test 3",
			Repository: []string{
				"https://username:password@example.com/",
				"registry",
			},
			ExpectedResult: "https://username:password@example.com/registry",
		},
		{

			Name: "test 4",
			Repository: []string{
				"http://username:password@example.com/",
				"registry",
			},
			ExpectedResult: "http://username:password@example.com/registry",
		},
		{

			Name: "test 5",
			Repository: []string{
				"username:password@example.com/",
				"registry",
			},
			ExpectedResult: "https://username:password@example.com/registry",
		},
		{
			Name: "test 6",
			Repository: []string{
				"username:password@example.com",
			},
			ExpectedResult: "https://username:password@example.com",
		},
	}

	for _, tt := range testData {
		t.Run(tt.Name, func(t *testing.T) {
			gotResult, gotError := joinURL(tt.Repository)

			if tt.WantErr {
				require.Error(t, gotError)
			} else {
				require.NoError(t, gotError)
			}

			assert.Equal(t, tt.ExpectedResult, gotResult)
		})
	}

}

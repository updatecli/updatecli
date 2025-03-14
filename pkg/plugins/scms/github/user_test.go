package github

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserInfo(t *testing.T) {

	testdata := []struct {
		user                 string
		expectedID           string
		expectedError        bool
		expectedErrorMessage string
	}{
		{
			user:       "olblak",
			expectedID: "MDQ6VXNlcjIzNjAyMjQ=",
		},
		{
			// Trying to get a organization instead of a user should fail
			user:                 "updatecli",
			expectedError:        true,
			expectedErrorMessage: "Could not resolve to a User with the login of 'updatecli'.",
		},
	}

	token := os.Getenv("GITHUB_TOKEN")

	if token == "" {
		t.Skip("GITHUB_TOKEN is not set so we can't test on the GitHub api")
	}

	for _, tt := range testdata {

		// Create a new instance of the Github plugin
		g, err := New(Spec{
			Owner:      "updatecli-test",
			Repository: "updatecli",
			Token:      token,
		}, "")
		require.NoError(t, err)

		// Call the GetUser function with a specific username
		gotUserInfo, err := getUserInfo(g.client, tt.user)

		if tt.expectedError {
			assert.Equal(t, tt.expectedErrorMessage, err.Error())
			require.Nil(t, gotUserInfo)
		} else {
			require.NoError(t, err)
			assert.Equal(t, tt.expectedID, gotUserInfo.ID)
		}

	}

}

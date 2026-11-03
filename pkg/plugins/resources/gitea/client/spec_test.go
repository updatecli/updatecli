package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/credential"
)

func TestSanitizeCredentials(t *testing.T) {
	testData := []struct {
		name             string
		spec             Spec
		envUsername      string
		envToken         string
		expectedUsername string
		expectedToken    string
	}{
		{
			name:             "Environment fills unset credentials",
			envUsername:      "env-user",
			envToken:         "env-token",
			expectedUsername: "env-user",
			expectedToken:    "env-token",
		},
		{
			name:             "Manifest wins over the environment",
			spec:             Spec{Username: "spec-user", Token: "spec-token"},
			envUsername:      "env-user",
			envToken:         "env-token",
			expectedUsername: "spec-user",
			expectedToken:    "spec-token",
		},
		{
			name:             "Environment token without a username",
			envToken:         "env-token",
			expectedUsername: defaultTokenUsername,
			expectedToken:    "env-token",
		},
		{
			name:             "Manifest token without a username",
			spec:             Spec{Token: "spec-token"},
			expectedUsername: defaultTokenUsername,
			expectedToken:    "spec-token",
		},
		{
			name: "No credential keeps anonymous access",
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(credential.ENVIRONMENT_GITEA_USERNAME, tt.envUsername)
			t.Setenv(credential.ENVIRONMENT_GITEA_TOKEN, tt.envToken)

			spec := tt.spec
			spec.URL = "gitea.example.com"

			require.NoError(t, spec.Sanitize())

			assert.Equal(t, tt.expectedUsername, spec.Username)
			assert.Equal(t, tt.expectedToken, spec.Token)
		})
	}
}

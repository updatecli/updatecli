package updatecli

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/compose"
)

func TestSearchFiles(t *testing.T) {

	gotFiles, err := searchUpdatecliComposeFiles(
		"testdata/website", DefaultFiles[:])
	if err != nil {
		t.Errorf("%s\n", err)
	}
	expectedFile := "testdata/website/updatecli-compose.yaml"

	if len(gotFiles) == 0 {
		t.Errorf("Expecting file %q but got none", expectedFile)
		return
	}

	if gotFiles[0] != expectedFile {
		t.Errorf("Expecting file %q but got %q", expectedFile, gotFiles[0])
	}
}

func TestListUpdatecliPolicies(t *testing.T) {

	gotMetadata, err := getComposeFileMetadata(
		"testdata/website/updatecli-compose.yaml")
	if err != nil {
		t.Errorf("%s\n", err)
	}
	expectedPolicies := []compose.Policy{
		{
			Name: "Local Updatecli Website Policies",
			Config: []string{
				"updatecli/updatecli.d/",
			},
		},
		{
			Name:   "Handle Nodejs version in githubaction",
			Policy: "ghcr.io/updatecli/policies/policies/nodejs/githubaction:latest",
			Values: []string{
				"updatecli/values.d/scm.yaml",
				"updatecli/values.d/nodejs.yaml",
			},
		},
		{
			Name:   "Handle Nodejs version in Netlify",
			Policy: "ghcr.io/updatecli/policies/policies/nodejs/netlify:0.1.0",
			Values: []string{
				"updatecli/values.d/scm.yaml",
				"updatecli/values.d/nodejs.yaml",
			},
		},
		{
			Name:   "Handle Hugo version in Netlify",
			Policy: "ghcr.io/updatecli/policies/policies/hugo/netlify:0.4.0@sha256:353d6cf2eb909c50bdb8d088f0df8ef53b0f90aec725a7a0c2b75ebe8d3352c1",
			Values: []string{
				"updatecli/values.d/scm.yaml",
			},
		},
	}

	require.Equal(t, expectedPolicies, gotMetadata.Policies)
}

func TestGetPolicyName(t *testing.T) {
	testdata := []struct {
		name            string
		expectedName    string
		expectedVersion string
		expectedErr     bool
		expectedErrMsg  string
	}{
		{
			name:            "ghcr.io/updatecli/policies/hugo/netlify:latest",
			expectedName:    "ghcr.io/updatecli/policies/hugo/netlify",
			expectedVersion: "latest",
		},
		{
			name:           "",
			expectedErr:    true,
			expectedErrMsg: "policy name is empty",
		},
		{
			name:            "ghcr.io/updatecli/policies/hugo/netlify:0.5.0@sha256:121231231",
			expectedName:    "ghcr.io/updatecli/policies/hugo/netlify",
			expectedVersion: "0.5.0",
		},
		{
			name:            "ghcr.io/updatecli/policies/hugo/netlify:0.5.0@sha256",
			expectedName:    "ghcr.io/updatecli/policies/hugo/netlify",
			expectedVersion: "0.5.0",
		},
	}

	for _, tt := range testdata {

		t.Run(tt.name, func(t *testing.T) {

			gotName, gotVersion, gotErr := getPolicyName(tt.name)

			if tt.expectedErr {
				require.Equal(t, gotErr.Error(), tt.expectedErrMsg)
			}

			require.Equal(t, tt.expectedName, gotName)
			require.Equal(t, tt.expectedVersion, gotVersion)
		})
	}
}

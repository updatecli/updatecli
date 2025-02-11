package githubaction

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchWorkflowFiles(t *testing.T) {
	g, err := New(
		Spec{
			RootDir: "testdata",
		}, "", "", "")

	require.NoError(t, err)

	err = g.searchWorkflowFiles("testdata", defaultWorkflowFiles[:])
	if err != nil {
		t.Error(err)
	}

	expectedWorkflowFiles := []string{
		"testdata/digest/.github/workflows/updatecli.yaml",
		"testdata/docker/.github/workflows/docker-01.yaml",
		"testdata/docker/.github/workflows/docker-02.yaml",
		"testdata/duplicate_steps/.github/workflows/updatecli.yaml",
		"testdata/gitea/.gitea/workflows/updatecli.yaml",
		"testdata/updatecli/.github/workflows/updatecli.yaml",
	}

	assert.Equal(t, expectedWorkflowFiles, g.workflowFiles)
}

func TestParseActionName(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		expectedURL        string
		expectedOwner      string
		expectedRepository string
		expectedReference  string
		expectedDirectory  string
		expectedActionKind string
	}{
		{
			name:               "complete action name",
			input:              "owner/repo@v1",
			expectedOwner:      "owner",
			expectedRepository: "repo",
			expectedReference:  "v1",
			expectedActionKind: ACTIONKINDDEFAULT,
		},
		{
			name:               "complete action name without reference",
			input:              "owner/repo",
			expectedOwner:      "owner",
			expectedRepository: "repo",
			expectedActionKind: ACTIONKINDDEFAULT,
		},
		{
			name:               "incomplete action name",
			input:              "owner@v1",
			expectedActionKind: "",
		},
		{
			name:               "GitHub url action",
			input:              "https://github.com/actions/checkout@v4",
			expectedURL:        "https://github.com",
			expectedOwner:      "actions",
			expectedRepository: "checkout",
			expectedReference:  "v4",
			expectedActionKind: ACTIONKINDDEFAULT,
		},
		{
			name:               "GitHub url action without scheme",
			input:              "github.com/actions/checkout@v4",
			expectedURL:        "github.com",
			expectedOwner:      "actions",
			expectedRepository: "checkout",
			expectedReference:  "v4",
			expectedActionKind: ACTIONKINDDEFAULT,
		},
		{
			name:               "Gitea url action",
			input:              "http://your_gitea.com/owner/repo@branch",
			expectedURL:        "http://your_gitea.com",
			expectedOwner:      "owner",
			expectedRepository: "repo",
			expectedReference:  "branch",
			expectedActionKind: ACTIONKINDDEFAULT,
		},
		{
			name:               "GitHub action with subdirectory",
			input:              "anchore/sbom-action/download-syft@v1",
			expectedOwner:      "anchore",
			expectedRepository: "sbom-action",
			expectedReference:  "v1",
			expectedDirectory:  "download-syft",
			expectedActionKind: ACTIONKINDDEFAULT,
		},
		{
			name:               "GitHub action with subdirectory without reference",
			input:              "anchore/sbom-action/download-syft",
			expectedOwner:      "anchore",
			expectedRepository: "sbom-action",
			expectedDirectory:  "download-syft",
			expectedActionKind: ACTIONKINDDEFAULT,
		},
		{
			name:               "GitHub action using a local path",
			input:              "./actions/checkout",
			expectedOwner:      "",
			expectedURL:        "",
			expectedRepository: "",
			expectedReference:  "",
			expectedDirectory:  "",
			expectedActionKind: ACTIONKINDLOCAL,
		},
		{
			name:               "Docker action",
			input:              "docker://alpine:latest",
			expectedOwner:      "",
			expectedURL:        "alpine:latest",
			expectedRepository: "",
			expectedReference:  "",
			expectedDirectory:  "",
			expectedActionKind: ACTIONKINDDOCKER,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, gotOwner, gotRepository, gotDirectory, gotReference, gotKind := parseActionName(tt.input)
			assert.Equal(t, tt.expectedURL, gotURL)
			assert.Equal(t, tt.expectedOwner, gotOwner)
			assert.Equal(t, tt.expectedRepository, gotRepository)
			assert.Equal(t, tt.expectedReference, gotReference)
			assert.Equal(t, tt.expectedDirectory, gotDirectory)
			assert.Equal(t, tt.expectedActionKind, gotKind)
		})
	}
}

func TestParseActionDigestComment(t *testing.T) {
	tests := []struct {
		name                    string
		input                   string
		expectedDigestReference string
	}{
		{
			name:                    "complete digest commit comment",
			input:                   "8f4b7f84864484a7bf31766abe9204da3cbe65b3",
			expectedDigestReference: "8f4b7f84864484a7bf31766abe9204da3cbe65b3",
		},
		{
			name:                    "complete digest tag comment",
			input:                   "v4.3.2 by",
			expectedDigestReference: "v4.3.2",
		},
		{
			name:                    "irrelevant comment",
			input:                   "This is a comment irrelevant",
			expectedDigestReference: "This",
		},
		{
			name:                    "empty digest comment",
			input:                   "",
			expectedDigestReference: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDigestReference := parseActionDigestComment(tt.input)
			assert.Equal(t, tt.expectedDigestReference, gotDigestReference)
		})
	}
}

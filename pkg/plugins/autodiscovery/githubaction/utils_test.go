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
		}, "", "")

	require.NoError(t, err)

	err = g.searchWorkflowFiles("testdata", defaultWorkflowFiles[:])
	if err != nil {
		t.Error(err)
	}

	expectedWorkflowFiles := []string{
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
	}{
		{
			name:               "complete action name",
			input:              "owner/repo@v1",
			expectedOwner:      "owner",
			expectedRepository: "repo",
			expectedReference:  "v1",
		},
		{
			name:               "complete action name without reference",
			input:              "owner/repo",
			expectedOwner:      "owner",
			expectedRepository: "repo",
		},
		{
			name:  "incomplete action name",
			input: "owner@v1",
		},
		{
			name:               "GitHub url action",
			input:              "https://github.com/actions/checkout@v4",
			expectedURL:        "https://github.com",
			expectedOwner:      "actions",
			expectedRepository: "checkout",
			expectedReference:  "v4",
		},
		{
			name:               "GitHub url action without scheme",
			input:              "github.com/actions/checkout@v4",
			expectedURL:        "github.com",
			expectedOwner:      "actions",
			expectedRepository: "checkout",
			expectedReference:  "v4",
		},
		{
			name:               "Gitea url action",
			input:              "http://your_gitea.com/owner/repo@branch",
			expectedURL:        "http://your_gitea.com",
			expectedOwner:      "owner",
			expectedRepository: "repo",
			expectedReference:  "branch",
		},
		{
			name:               "GitHub action with subdirectory",
			input:              "anchore/sbom-action/download-syft@v1",
			expectedOwner:      "anchore",
			expectedRepository: "sbom-action",
			expectedReference:  "v1",
			expectedDirectory:  "download-syft",
		},
		{
			name:               "GitHub action with subdirectory without reference",
			input:              "anchore/sbom-action/download-syft",
			expectedOwner:      "anchore",
			expectedRepository: "sbom-action",
			expectedDirectory:  "download-syft",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, gotOwner, gotRepository, gotDirectory, gotReference := parseActionName(tt.input)
			assert.Equal(t, tt.expectedURL, gotURL)
			assert.Equal(t, tt.expectedOwner, gotOwner)
			assert.Equal(t, tt.expectedRepository, gotRepository)
			assert.Equal(t, tt.expectedReference, gotReference)
			assert.Equal(t, tt.expectedDirectory, gotDirectory)
		})
	}
}

package githubaction

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadGitHubActionWorkflow(t *testing.T) {
	testdata := []struct {
		name                  string
		stepName              string
		stepUses              string
		expectedStepName      string
		expectedStepUses      string
		expectedCommentDigest string
	}{
		{
			name:                  "Reference specific commit",
			stepName:              "Checkout",
			stepUses:              "actions/checkout@8f4b7f84864484a7bf31766abe9204da3cbe65b3",
			expectedStepName:      "Checkout",
			expectedStepUses:      "actions/checkout@8f4b7f84864484a7bf31766abe9204da3cbe65b3",
			expectedCommentDigest: "",
		},
		{
			name:                  "Reference specific commit with pinned digest",
			stepName:              "Checkout",
			stepUses:              "actions/checkout@8f4b7f84864484a7bf31766abe9204da3cbe65b3  # pinned from 8f4b7f84864484a7bf31766abe9204da3cbe65b3 by updatecli (do-not-remove-comment)",
			expectedStepName:      "Checkout",
			expectedStepUses:      "actions/checkout@8f4b7f84864484a7bf31766abe9204da3cbe65b3",
			expectedCommentDigest: "pinned from 8f4b7f84864484a7bf31766abe9204da3cbe65b3 by updatecli (do-not-remove-comment)",
		},
		{
			name:                  "Reference major version",
			stepName:              "Checkout",
			stepUses:              "actions/checkout@v4",
			expectedStepName:      "Checkout",
			expectedStepUses:      "actions/checkout@v4",
			expectedCommentDigest: "",
		},
		{
			name:                  "Reference major version with pinned digest",
			stepName:              "Checkout",
			stepUses:              "actions/checkout@v4  # pinned from v4 by updatecli (do-not-remove-comment)",
			expectedStepName:      "Checkout",
			expectedStepUses:      "actions/checkout@v4",
			expectedCommentDigest: "pinned from v4 by updatecli (do-not-remove-comment)",
		},
		{
			name:                  "Reference specific version",
			stepName:              "Checkout",
			stepUses:              "actions/checkout@v4.2.0",
			expectedStepName:      "Checkout",
			expectedStepUses:      "actions/checkout@v4.2.0",
			expectedCommentDigest: "",
		},
		{
			name:                  "Reference specific version with pinned digest",
			stepName:              "Checkout",
			stepUses:              "actions/checkout@v4.2.0  # pinned from v4.2.0 by updatecli (do-not-remove-comment)",
			expectedStepName:      "Checkout",
			expectedStepUses:      "actions/checkout@v4.2.0",
			expectedCommentDigest: "pinned from v4.2.0 by updatecli (do-not-remove-comment)",
		},
		{
			name:                  "Reference branch",
			stepName:              "Checkout",
			stepUses:              "actions/checkout@main",
			expectedStepName:      "Checkout",
			expectedStepUses:      "actions/checkout@main",
			expectedCommentDigest: "",
		},
		{
			name:                  "Reference branch with pinned digest",
			stepName:              "Checkout",
			stepUses:              "actions/checkout@main  # pinned from main by updatecli (do-not-remove-comment)",
			expectedStepName:      "Checkout",
			expectedStepUses:      "actions/checkout@main",
			expectedCommentDigest: "pinned from main by updatecli (do-not-remove-comment)",
		},
		{
			name:                  "Reference subdirectory in a GitHub repository",
			stepName:              "AWS EC2 Action",
			stepUses:              "actions/aws/ec2@main",
			expectedStepName:      "AWS EC2 Action",
			expectedStepUses:      "actions/aws/ec2@main",
			expectedCommentDigest: "",
		},
		{
			name:                  "Reference subdirectory in a GitHub repository with pinned digest",
			stepName:              "AWS EC2 Action",
			stepUses:              "actions/aws/ec2@main  # pinned from main by updatecli (do-not-remove-comment)",
			expectedStepName:      "AWS EC2 Action",
			expectedStepUses:      "actions/aws/ec2@main",
			expectedCommentDigest: "pinned from main by updatecli (do-not-remove-comment)",
		},
		{
			name:                  "Reference local action",
			stepName:              "Local Action",
			stepUses:              "./.github/actions/my-action",
			expectedStepName:      "Local Action",
			expectedStepUses:      "./.github/actions/my-action",
			expectedCommentDigest: "",
		},
		{
			name:                  "Reference Docker public registry action",
			stepName:              "Docker Gradle Action",
			stepUses:              "docker://gcr.io/cloud-builders/gradle",
			expectedStepName:      "Docker Gradle Action",
			expectedStepUses:      "docker://gcr.io/cloud-builders/gradle",
			expectedCommentDigest: "",
		},
		{
			name:                  "Reference Docker Hub image",
			stepName:              "Alpine Docker Image",
			stepUses:              "docker://alpine:3.8",
			expectedStepName:      "Alpine Docker Image",
			expectedStepUses:      "docker://alpine:3.8",
			expectedCommentDigest: "",
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			workflowContent := fmt.Sprintf(`name: Test Workflow

on:
  push:
    branches:
      - main

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: %s
        uses: %s`, tt.stepName, tt.stepUses)

			// Create a temp file to pass to the function
			tempFile, err := os.CreateTemp("", "workflow-*.yaml")
			require.NoError(t, err)
			defer os.Remove(tempFile.Name())

			_, err = tempFile.WriteString(workflowContent)
			require.NoError(t, err)
			tempFile.Close()

			// Load the workflow
			w, err := loadGitHubActionWorkflow(tempFile.Name())
			require.NoError(t, err)

			// Validate the step
			s := w.Jobs["build"].Steps[0]
			assert.Equal(t, tt.expectedStepName, s.Name)
			assert.Equal(t, tt.expectedStepUses, s.Uses)
			assert.Equal(t, tt.expectedCommentDigest, s.CommentDigest)
		})
	}
}

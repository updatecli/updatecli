package gitgeneric

import (
	"os"
	"path/filepath"
	"testing"
)

// Test that we can correctly retrieve a list of tags from a remote git repository
// and that it's correctly ordered, starting with the oldest tag
func TestBranchIntegration(t *testing.T) {
	g := GoGit{}
	workingDir := filepath.Join(os.TempDir(), "tests", "updatecli")
	withSubmodules := true
	err := g.Clone("", "", "https://github.com/updatecli/updatecli-action.git", workingDir, &withSubmodules)
	if err != nil {
		t.Errorf("Don't expect error: %q", err)
	}
	defer os.RemoveAll(workingDir)

	branches, err := g.Branches(workingDir)
	if err != nil {
		t.Errorf("Don't expect error: %q", err)
	}

	// Branch v1 is the oldest branch, and there is no plan to remove it
	expectedBranch := "v1"

	// Test that the first tag from array is also the oldest one
	if branches[0] != expectedBranch {
		t.Errorf("Expected tag %q to be found in %q", expectedBranch, branches)
	}
	os.Remove(workingDir)
}

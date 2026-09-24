package gitbranch

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

// recordingGitHandler records the directory every call is made against.
type recordingGitHandler struct {
	gitgeneric.GitHandler
	branchesDir   string
	newBranchDir  string
	checkoutDir   string
	pushBranchDir string
}

func (r *recordingGitHandler) Branches(workingDir string) ([]string, error) {
	r.branchesDir = workingDir
	return []string{"main"}, nil
}

func (r *recordingGitHandler) NewBranch(branch, workingDir string) (bool, error) {
	r.newBranchDir = workingDir
	return true, nil
}

func (r *recordingGitHandler) Checkout(username, password, branch, remoteBranch, workingDir string, forceReset bool, depth *int) error {
	r.checkoutDir = workingDir
	return nil
}

func (r *recordingGitHandler) PushBranch(branch, username, password, workingDir string, force bool) error {
	r.pushBranchDir = workingDir
	return nil
}

// TestGitBranch_TargetCreatesBranchInResolvedDirectory checks that the branch is created in
// the same repository the target inspected, which is spec.path resolved from the manifest
// directory, and not in spec.path as written.
func TestGitBranch_TargetCreatesBranchInResolvedDirectory(t *testing.T) {
	handler := &recordingGitHandler{}

	gb := GitBranch{
		spec: Spec{
			Path:         "repo",
			Branch:       "feature",
			SourceBranch: "main",
		},
		nativeGitHandler: handler,
	}

	gotResult := result.Target{}
	err := gb.Target(context.Background(), "", nil, pathresolver.New(nil, "updatecli.d"), false, &gotResult)
	require.NoError(t, err)

	expectedDir := filepath.Join("updatecli.d", "repo")
	assert.Equal(t, expectedDir, handler.branchesDir)
	assert.Equal(t, expectedDir, handler.newBranchDir)
	assert.Equal(t, expectedDir, handler.checkoutDir)
	assert.Equal(t, expectedDir, handler.pushBranchDir)
	assert.True(t, gotResult.Changed)
}

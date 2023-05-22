package gitgeneric

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Based on this discussion,
// https://github.com/updatecli/updatecli/pull/436#discussion_r777184192
// I decided to comment the follow tests so we can start using this fix
// func TestIsSimilarBranch(t *testing.T) {
//
// 	type data struct {
// 		branchA        string
// 		branchB        string
// 		workingDir     string
// 		expectedResult bool
// 		expectedError  error
// 	}
//
// 	type dataSet []data
//
// 	dSet := dataSet{
// 		{
// 			branchA:        "main",
// 			branchB:        "issue-285",
// 			workingDir:     "../../../..",
// 			expectedResult: false,
// 			expectedError:  nil,
// 		},
// 		{
// 			branchA:        "main",
// 			branchB:        "main",
// 			workingDir:     "../../../..",
// 			expectedResult: true,
// 			expectedError:  nil,
// 		},
// 		{
// 			branchA:        "main",
// 			branchB:        "doNotExist",
// 			workingDir:     "../../../..",
// 			expectedResult: false,
// 			expectedError:  fmt.Errorf("reference not found"),
// 		},
// 	}
//
// 	for id, d := range dSet {
// 		t.Run(fmt.Sprint(id), func(t *testing.T) {
// 			got, err := IsSimilarBranch(
// 				d.branchA,
// 				d.branchB,
// 				d.workingDir)
//
// 			if d.expectedError != nil {
// 				assert.Equal(t, err, d.expectedError)
// 				return
// 			}
//
// 			require.NoError(t, err)
// 			assert.Equal(t, got, d.expectedResult)
// 		})
// 	}
// }

func TestSanitizeBranchName(t *testing.T) {
	type dataSet struct {
		branch   string
		expected string
	}

	datasets := []dataSet{
		{
			branch:   "master",
			expected: "master",
		},
		{
			branch:   "master+",
			expected: "master",
		},
		{
			branch:   "mas:ter",
			expected: "master",
		},
		{
			branch:   "ma*ster+",
			expected: "master",
		},
		{
			branch:   "m:a*ster+",
			expected: "master",
		},
		{
			branch:   "m:a*st  er+",
			expected: "master",
		},
		{
			branch:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			expected: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
		{
			branch:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaabbb",
			expected: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
	}
	g := GoGit{}
	for _, d := range datasets {
		got := g.SanitizeBranchName(d.branch)
		if got != d.expected {
			t.Errorf("Branch name isn't correctly got %s, expected %s", got, d.expected)
		}
	}
}

// Test that we can correctly retrieve a list of tags from a remote git repository
// and that it's correctly ordered, starting with the oldest tag
func TestTagsIntegration(t *testing.T) {
	g := GoGit{}
	workingDir := filepath.Join(os.TempDir(), "tests", "updatecli")
	err := g.Clone("", "", "https://github.com/updatecli/updatecli.git", workingDir)
	if err != nil {
		t.Errorf("Don't expect error: %q", err)
	}
	defer os.RemoveAll(workingDir)

	tags, err := g.Tags(workingDir)
	if err != nil {
		t.Errorf("Don't expect error: %q", err)
	}

	expectedTag := "v0.0.1"
	found := false

	// Test that the first tag from array is also the oldest one
	if tags[0] == expectedTag {
		found = true
	}

	if !found {
		t.Errorf("Expected tag %q to be found in %q", expectedTag, tags)
	}
	os.Remove(workingDir)
}

// Test that we can correctly retrieve a list of tag refs from a remote git repository
// and that it's correctly ordered, starting with the oldest tag
func TestTagRefsIntegration(t *testing.T) {
	g := GoGit{}
	workingDir := filepath.Join(os.TempDir(), "tests", "updatecli")
	err := g.Clone("", "", "https://github.com/updatecli/updatecli.git", workingDir)
	if err != nil {
		t.Errorf("Don't expect error: %q", err)
	}
	defer os.RemoveAll(workingDir)

	refs, err := g.TagRefs(workingDir)
	if err != nil {
		t.Errorf("Don't expect error: %q", err)
	}

	expectedRef := DatedTag{
		When: time.Unix(1582144213, 0).In(time.FixedZone("", 1*60*60*1)), //tz "+0100"
		Name: "v0.0.1",
		Hash: "d0812d972468d97a3b7e70699f977854cfb83892",
	}
	found := false

	// Test that the first tag from array is also the oldest one
	if refs[0] == expectedRef {
		found = true
	}

	if !found {
		t.Errorf("Expected ref %v to be %v", expectedRef, refs[0])
	}
	os.Remove(workingDir)
}

// Test that we can correctly retrieve tag hashes from a remote git repository
// and that it's correctly ordered, starting with the oldest tag
func TestHashesIntegration(t *testing.T) {
	g := GoGit{}
	workingDir := filepath.Join(os.TempDir(), "tests", "updatecli")
	err := g.Clone("", "", "https://github.com/updatecli/updatecli.git", workingDir)
	if err != nil {
		t.Errorf("Don't expect error: %q", err)
	}
	defer os.RemoveAll(workingDir)

	hashes, err := g.TagHashes(workingDir)
	if err != nil {
		t.Errorf("Don't expect error: %q", err)
	}

	expectedHash := "d0812d972468d97a3b7e70699f977854cfb83892"
	found := false

	// Test that the first tag from array is also the oldest one
	if hashes[0] == expectedHash {
		found = true
	}

	if !found {
		t.Errorf("Expected tag %q to be found in %q", expectedHash, hashes)
	}
	os.Remove(workingDir)
}
func TestGoGit_RemoteURLs(t *testing.T) {
	cwd, _ := os.Getwd()

	tests := []struct {
		name       string
		workingDir string
		wantErr    bool
	}{
		{
			name:       "passing test with the current working directory",
			workingDir: cwd,
		},
		{
			name:       "failing test with existing directory with no git in it",
			workingDir: "/tmp",
			wantErr:    true,
		},
		{
			name:       "failing test with non-existing directory",
			workingDir: "/nonexistent",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := GoGit{}
			gotRemotes, gotErr := g.RemoteURLs(tt.workingDir)

			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			// Only testing that there is an "origin" remote with a non empty URL.
			// Because origin's URL, as well as other remotes, depends on the user or CI configuration and is not deterministic.
			assert.Contains(t, gotRemotes, "origin")
			assert.NotEmpty(t, gotRemotes["origin"])
		})
	}
}

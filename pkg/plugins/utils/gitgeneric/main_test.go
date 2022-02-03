package gitgeneric

import (
	"os"
	"path/filepath"
	"testing"
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

	for _, d := range datasets {
		got := SanitizeBranchName(d.branch)
		if got != d.expected {
			t.Errorf("Branch name isn't correctly got %s, expected %s", got, d.expected)
		}
	}
}

// Test that we can correctly retrieve a list of tags from a remote git repository
// and that it's correctly ordered, starting with the oldest tag
func TestTagsIntegration(t *testing.T) {

	workingDir := filepath.Join(os.TempDir(), "tests", "updatecli")
	err := Clone("", "", "https://github.com/updatecli/updatecli.git", workingDir)
	if err != nil {
		t.Errorf("Don't expect error: %q", err)
	}
	defer os.RemoveAll(workingDir)

	tags, err := Tags(workingDir)
	if err != nil {
		t.Errorf("Don't expect error: %q", err)
	}

	expectedTag := "untagged-902d9ce264ba6334c5d0"
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

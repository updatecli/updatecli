package gitgeneric

import (
	"os"
	"path/filepath"
	"strings"
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
			expected: "master_cc04da33a81fe331803ca1f1cd5cdfd963f4f1e04e467f11480a633f4cced003",
		},
		{
			branch:   "mas:ter",
			expected: "master_05ef7708208c5918a3e64e7ec5e3d0dcd3797343768093d456496ae91ace415a",
		},
		{
			branch:   "ma*ster+",
			expected: "master_0ebbd72ba6189c8cbd3cca110d81c7462d384b086c39e41698c4987c388ca1b0",
		},
		{
			branch:   "m:a*ster+",
			expected: "master_e2dae9cccecaf7aa10959d3496c25647ba56f44fe2219a5246615e6eab0b3619",
		},
		{
			branch:   "m:a*st  er+",
			expected: "master_4e3c17615bfe313eb80c25572c362d0ee21b9c01390aa065806fd14b57ad36b2",
		},
		{
			branch:   strings.Repeat("a", 255),
			expected: strings.Repeat("a", 255),
		},
		{
			branch:   strings.Repeat("a", 254) + "+",
			expected: strings.Repeat("a", 190) + "_f496626d3405726ec9be384284f8742913b14bfbc2dbd529199ae9acad05eef1",
		},
		{
			branch:   strings.Repeat("a", 255) + "bbb",
			expected: strings.Repeat("a", 190) + "_cf09e931ae8bac274929c6ff633a8fdd8dd19f0d7f723af1f86227eb53cbb548",
		},
		{
			branch:   strings.Repeat("a", 255) + "ccc",
			expected: strings.Repeat("a", 190) + "_22f109b028d0ee7d1d608ded9d670d71147d5e1ab4a32d1f2137e49fb0d48b85",
		},
		{
			branch:   "updatecli_release+1_dependencies",
			expected: "updatecli_release1_dependencies_41656c7604d017869bf9096e4b7b4468a32eaa98d3395416491d1be63c7ade90",
		},
		{
			branch:   "updatecli_release1_dependencies",
			expected: "updatecli_release1_dependencies",
		},
		{
			branch:   "updatecli_release{1}_dependencies",
			expected: "updatecli_release_1__dependencies_7d31fd438a021bb1f89bacf7cfb0040d6e99da23c14d6a1cadffd3fdf06fe98a",
		},
		{
			branch:   "updatecli_release_1__dependencies",
			expected: "updatecli_release_1__dependencies",
		},
		{
			branch:   "/updatecli_main_dependencies",
			expected: "updatecli_main_dependencies_1d2198a5d46e487ff069d53d58db76438ef23ec923a540521e619c8fde66016e",
		},
		{
			branch:   "updatecli_release/1.x_dependencies",
			expected: "updatecli_release/1.x_dependencies",
		},
		{
			branch:   strings.Repeat("a", 189) + "\u00e9" + strings.Repeat("b", 65),
			expected: strings.Repeat("a", 189) + "_0c81f0710be09e36ae00000e5b823cc4e16e884c4c3a491f06b56ec991cdcd7e",
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
	withSubmodules := true
	err := g.Clone("", "", "https://github.com/updatecli/updatecli.git", workingDir, &withSubmodules, nil, "", false)
	if err != nil {
		t.Errorf("Don't expect error: %q", err)
	}
	defer os.RemoveAll(workingDir)

	tags, err := g.Tags(workingDir)
	if err != nil {
		t.Errorf("Don't expect error: %q", err)
	}

	expectedTag := "v0.0.1"

	// Test that the first tag from array is also the oldest one
	if tags[0] != expectedTag {
		t.Errorf("Expected tag %q to be found in %q", expectedTag, tags)
	}
	os.Remove(workingDir)
}

// Test that we can correctly retrieve a list of tag refs from a remote git repository
// and that it's correctly ordered, starting with the oldest tag
func TestTagRefsIntegration(t *testing.T) {
	g := GoGit{}
	workingDir := filepath.Join(os.TempDir(), "tests", "updatecli")
	withSubmodules := true
	err := g.Clone("", "", "https://github.com/updatecli/updatecli.git", workingDir, &withSubmodules, nil, "", false)
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

	// Test that the first tag from array is also the oldest one
	if refs[0] != expectedRef {
		t.Errorf("Expected ref %v to be %v", expectedRef, refs[0])
	}
	os.Remove(workingDir)
}

// Test that we can correctly retrieve tag hashes from a remote git repository
// and that it's correctly ordered, starting with the oldest tag
func TestHashesIntegration(t *testing.T) {
	g := GoGit{}
	workingDir := filepath.Join(os.TempDir(), "tests", "updatecli")
	withSubmodules := true
	err := g.Clone("", "", "https://github.com/updatecli/updatecli.git", workingDir, &withSubmodules, nil, "", false)
	if err != nil {
		t.Errorf("Don't expect error: %q", err)
	}
	defer os.RemoveAll(workingDir)

	hashes, err := g.TagHashes(workingDir)
	if err != nil {
		t.Errorf("Don't expect error: %q", err)
	}

	expectedHash := "d0812d972468d97a3b7e70699f977854cfb83892"

	if hashes[0] != expectedHash {
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

// Test that we can correctly retrieve submodule content from a remote git repository
func TestSubmodulesEnabledContent(t *testing.T) {
	g := GoGit{}
	workingDir := filepath.Join(os.TempDir(), "tests", "updatecli-submodules")
	withSubmodules := true
	err := g.Clone("", "", "https://github.com/updatecli-test/updatecli-submodules.git", workingDir, &withSubmodules, nil, "", false)
	if err != nil {
		t.Errorf("Don't expect error: %q", err)
	}
	defer os.RemoveAll(workingDir)

	checkFiles := []string{
		"nocode/README.md",
		"cargo-lab/Cargo.toml",
	}

	for _, file := range checkFiles {
		assert.FileExists(t, filepath.Join(workingDir, file))
	}

	os.Remove(workingDir)
}

// Test that we can correctly retrieve submodule content from a remote git repository
func TestSubmodulesDisabledContent(t *testing.T) {
	g := GoGit{}
	workingDir := filepath.Join(os.TempDir(), "tests", "updatecli-submodules")
	withSubmodules := false
	err := g.Clone("", "", "https://github.com/updatecli-test/updatecli-submodules.git", workingDir, &withSubmodules, nil, "", false)
	if err != nil {
		t.Errorf("Don't expect error: %q", err)
	}
	defer os.RemoveAll(workingDir)

	checkFiles := []string{
		"nocode/README.md",
		"cargo-lab/Cargo.toml",
	}

	for _, file := range checkFiles {
		assert.NoFileExists(t, filepath.Join(workingDir, file))
	}

	os.Remove(workingDir)
}

func TestLatestCommitHash(t *testing.T) {
	g := GoGit{}
	workingDir := filepath.Join(os.TempDir(), "tests", "updatecli-submodules")
	withSubmodules := false

	tests := []struct {
		name       string
		cloneUrl   string
		workingDir string
		wantErr    bool
	}{
		{
			name:       "passing test with cloned repository",
			workingDir: workingDir,
			cloneUrl:   "https://github.com/updatecli-test/updatecli-submodules.git",
			wantErr:    false,
		},
		{
			name:       "failing test with existing directory with no git in it",
			workingDir: "/nonexistent/tmp/dir",
			cloneUrl:   "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer os.RemoveAll(workingDir)
			if tt.cloneUrl != "" {
				err := g.Clone("", "", tt.cloneUrl, tt.workingDir, &withSubmodules, nil, "", false)
				if err != nil {
					t.Errorf("unexpected error: %q", err)
				}
			}
			hash, err := g.GetLatestCommitHash(workingDir)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, hash, 40)
		})
	}
}

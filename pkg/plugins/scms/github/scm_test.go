package github

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommit(t *testing.T) {
	tests := []struct {
		name          string
		spec          Spec
		commitMsg     string
		mockedQuery   *commitQuery
		mockedError   error
		wantChangelog string
		wantErr       bool
	}{
		{
			name: "Case with error returned from query",
			spec: Spec{
				Owner:      "updatecli",
				Repository: "updatecli",
				Username:   "joe",
				Token:      "SuperSecretToken",
			},
			commitMsg:   "test commit",
			mockedQuery: &commitQuery{},
			mockedError: fmt.Errorf("Dummy error from github.com."),
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.mockedQuery)

			sut, err := New(tt.spec, "id1")

			require.NoError(t, err)

			sut.client = &MockGitHubClient{
				mockedQuery: tt.mockedQuery,
				mockedErr:   tt.mockedError,
			}

			_, err = sut.CreateCommit(context.Background(), tt.spec.Directory, tt.commitMsg, 0)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func base64Content(content string) githubv4.Base64String {
	return githubv4.Base64String(base64.StdEncoding.EncodeToString([]byte(content)))
}

func TestProcessChangedFiles(t *testing.T) {
	tests := []struct {
		name          string
		createFiles   []string
		files         []string
		want          []githubv4.FileAddition
		wantDeletions []githubv4.FileDeletion
		wantErr       bool
	}{
		{
			name:        "valid files",
			createFiles: []string{"file1.txt", "file2.txt"},
			files:       []string{"file1.txt", "file2.txt"},
			want: []githubv4.FileAddition{
				{
					Path:     githubv4.String("file1.txt"),
					Contents: base64Content("test content0"),
				},
				{
					Path:     githubv4.String("file2.txt"),
					Contents: base64Content("test content1"),
				},
			},
			wantDeletions: []githubv4.FileDeletion{},
		},
		{
			name:          "missing file is a deletion",
			createFiles:   nil,
			files:         []string{"file1.txt"},
			want:          []githubv4.FileAddition{},
			wantDeletions: []githubv4.FileDeletion{{Path: githubv4.String("file1.txt")}},
		},
		{
			name:        "mixed additions and deletions",
			createFiles: []string{"kept.txt"},
			files:       []string{"kept.txt", "removed.txt"},
			want: []githubv4.FileAddition{
				{
					Path:     githubv4.String("kept.txt"),
					Contents: base64Content("test content0"),
				},
			},
			wantDeletions: []githubv4.FileDeletion{{Path: githubv4.String("removed.txt")}},
		},
		{
			name:          "deleted file in a subdirectory",
			createFiles:   nil,
			files:         []string{"charts/sub/old.tgz"},
			want:          []githubv4.FileAddition{},
			wantDeletions: []githubv4.FileDeletion{{Path: githubv4.String("charts/sub/old.tgz")}},
		},
		{
			name:          "deleted file under a path replaced by a regular file",
			createFiles:   []string{"charts"},
			files:         []string{"charts/sub/old.tgz"},
			want:          []githubv4.FileAddition{},
			wantDeletions: []githubv4.FileDeletion{{Path: githubv4.String("charts/sub/old.tgz")}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "")
			require.NoError(t, err)

			defer os.RemoveAll(tempDir)

			for i, file := range tt.createFiles {
				fullPath := filepath.Join(tempDir, file)
				require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))

				tempFile, err := os.Create(fullPath)
				require.NoError(t, err)
				defer os.Remove(tempFile.Name())
				_, err = tempFile.WriteString("test content" + strconv.Itoa(i))
				require.NoError(t, err)
				require.NoError(t, tempFile.Close())
			}

			gotAdditions, gotDeletions, err := processChangedFiles(tempDir, tt.files)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, gotAdditions)
			assert.Equal(t, tt.wantDeletions, gotDeletions)
		})
	}
}

func TestProcessChangedFilesUnreadableFile(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("file permissions are not enforced on Windows or when running as root")
	}

	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	fullPath := filepath.Join(tempDir, "unreadable.txt")
	require.NoError(t, os.WriteFile(fullPath, []byte("test content"), 0o600))
	require.NoError(t, os.Chmod(fullPath, 0o000))
	defer os.Chmod(fullPath, 0o600) //nolint:errcheck

	_, _, err = processChangedFiles(tempDir, []string{"unreadable.txt"})
	require.Error(t, err)
}

func TestProcessChangedFilesDanglingSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation needs elevated privileges on Windows")
	}

	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	linkPath := filepath.Join(tempDir, "link")
	require.NoError(t, os.Symlink(filepath.Join(tempDir, "missing-target"), linkPath))

	_, _, err = processChangedFiles(tempDir, []string{"link"})
	require.Error(t, err)
}

func TestProcessChangedFilesSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation needs elevated privileges on Windows")
	}

	tempDir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	targetPath := filepath.Join(tempDir, "target.txt")
	require.NoError(t, os.WriteFile(targetPath, []byte("test content"), 0o600))

	linkPath := filepath.Join(tempDir, "link")
	require.NoError(t, os.Symlink(targetPath, linkPath))

	_, _, err = processChangedFiles(tempDir, []string{"link"})
	require.Error(t, err)
}

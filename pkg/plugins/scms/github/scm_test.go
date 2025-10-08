package github

import (
	"fmt"
	"os"
	"path/filepath"
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

			_, err = sut.CreateCommit(tt.spec.Directory, tt.commitMsg, 0)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestProcessChangedFiles(t *testing.T) {
	tests := []struct {
		name    string
		create  bool
		files   []string
		want    []githubv4.FileAddition
		wantErr bool
	}{
		{
			name:   "Case with valid files",
			create: true,
			files:  []string{"file1.txt", "file2.txt"},
			want: []githubv4.FileAddition{
				{
					Path:     githubv4.String("file1.txt"),
					Contents: githubv4.Base64String("dGVzdCBjb250ZW50MA=="), // no-spell-check-line
				},
				{
					Path:     githubv4.String("file2.txt"),
					Contents: githubv4.Base64String("dGVzdCBjb250ZW50MQ=="), // no-spell-check-line
				},
			},
			wantErr: false,
		},
		{
			name:    "Case with error encoding file",
			create:  false,
			files:   []string{"file1.txt"},
			want:    []githubv4.FileAddition{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "")
			require.NoError(t, err)

			defer os.RemoveAll(tempDir)

			if tt.create {
				for i, file := range tt.files {
					tempFile, err := os.Create(filepath.Join(tempDir, filepath.Base(file)))
					require.NoError(t, err)
					defer os.Remove(tempFile.Name())
					_, err = tempFile.WriteString("test content" + strconv.Itoa(i))
					require.NoError(t, err)
				}
			}

			got, err := processChangedFiles(tempDir, tt.files)

			if tt.wantErr {
				assert.Error(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

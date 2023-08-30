package gitgeneric

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadFileFromBranch(t *testing.T) {

	testCases := []struct {
		name        string
		repoDir     string
		revision    string
		filePath    string
		expectedErr error
		wantErr     bool
	}{
		{
			name:     "Read README.adoc from tag v0.20.0 branch",
			repoDir:  "../../../../",
			revision: "v0.20.0",
			filePath: "README.adoc",
		},
		{
			name:        "Read README.adoc from nonexistent tag v0.0.42",
			repoDir:     "../../../../",
			revision:    "v0.0.42",
			filePath:    "README.adoc",
			expectedErr: fmt.Errorf(`resolve revision "v0.0.42": reference not found`),
			wantErr:     true,
		},
		{
			name:        "Read doNotExist.adoc from tag v0.20.0 branch",
			repoDir:     "../../../../",
			revision:    "v0.20.0",
			filePath:    "doNotExist.adoc",
			expectedErr: fmt.Errorf("file not found"),
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ReadFileFromRevision(tc.repoDir, tc.revision, tc.filePath)

			switch tc.wantErr {
			case true:
				require.Equal(t, err.Error(), tc.expectedErr.Error())
			case false:
				require.NoError(t, err)
			}
		})
	}
}

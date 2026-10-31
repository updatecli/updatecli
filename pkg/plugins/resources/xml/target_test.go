package xml

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

func TestTarget(t *testing.T) {

	testData := []struct {
		name             string
		spec             Spec
		expectedResult   bool
		expectedErrorMsg error
		wantErr          bool
	}{
		{
			name: "Test 1",
			spec: Spec{
				File:  "testdata/data_0.xml",
				Path:  "/name/firstname",
				Value: "Bob",
			},
			expectedResult: true,
		},
		{
			name: "Test 2",
			spec: Spec{
				File:  "testdata/data_0.xml",
				Path:  "/name/firstname",
				Value: "John",
			},
			expectedResult: false,
		},
		{
			name: "Test 3",
			spec: Spec{
				File:  "testdata/data_2.xml",
				Path:  "/name/firstname",
				Value: "Bob",
			},
			expectedResult: true,
		},
		{
			name: "Test 4",
			spec: Spec{
				File:  "testdata/doNotExist.xml",
				Path:  "/name/firstname",
				Value: "Alice",
			},
			expectedResult:   false,
			expectedErrorMsg: errors.New("file \"testdata/doNotExist.xml\" does not exist"),
			wantErr:          true,
		},
		{
			name: "Test 5",
			spec: Spec{
				File:  "testdata/data_2.xml",
				Path:  "/name/nonexistent",
				Value: "Bob",
			},
			expectedResult:   false,
			wantErr:          true,
			expectedErrorMsg: errors.New("nothing found at path \"/name/nonexistent\" from file \"testdata/data_2.xml\""),
		},
		{
			name: "Test 6",
			spec: Spec{
				File:  "testdata/data_2.xml",
				Path:  "/name/firstname",
				Value: "John",
			},
			expectedResult: false,
		},
		{
			name: "Test 7",
			spec: Spec{
				File:  "https://raw.githubusercontent.com/updatecli/updatecli/main/pkg/plugins/resources/xml/testdata/data_2.xml",
				Path:  "/name/firstname",
				Value: "John",
			},
			wantErr:          true,
			expectedResult:   false,
			expectedErrorMsg: errors.New("URL scheme is not supported for XML target: \"https://raw.githubusercontent.com/updatecli/updatecli/main/pkg/plugins/resources/xml/testdata/data_2.xml\""),
		},
		{
			name: "Whitespace around a multi-line value is ignored",
			spec: Spec{
				File:  "testdata/data_3.xml",
				Path:  "/project/dependency/version",
				Value: "3.2.1",
			},
			expectedResult: false,
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {
			x, err := New(tt.spec)

			require.NoError(t, err)

			gotResult := result.Target{}
			err = x.Target(context.Background(), "", nil, pathresolver.Resolver{}, true, &gotResult)

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult.Changed)
		})
	}

}

func TestTargetKeepsMultiLineLayout(t *testing.T) {
	content, err := os.ReadFile("testdata/data_3.xml")
	require.NoError(t, err)

	file := filepath.Join(t.TempDir(), "pom.xml")
	require.NoError(t, os.WriteFile(file, content, 0o600)) //nolint:gosec // G703 false positive: file is under t.TempDir()

	x, err := New(Spec{File: file, Path: "/project/dependency/version", Value: "3.2.2"})
	require.NoError(t, err)

	gotResult := result.Target{}
	require.NoError(t, x.Target(context.Background(), "", nil, pathresolver.Resolver{}, false, &gotResult))

	assert.True(t, gotResult.Changed)
	assert.Equal(t, "3.2.1", gotResult.Information)

	got, err := os.ReadFile(file)
	require.NoError(t, err)
	// Only the version changes, the indentation around it is kept
	assert.Contains(t, string(got), "<version>\n      3.2.2\n    </version>")
}

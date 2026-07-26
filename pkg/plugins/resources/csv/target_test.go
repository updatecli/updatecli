package csv

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestTarget(t *testing.T) {

	testData := []struct {
		name             string
		spec             Spec
		sourceInput      string
		expectedResult   bool
		expectedErrorMsg error
		wantErr          bool
	}{
		{
			name: "Default successful workflow",
			spec: Spec{
				File: "testdata/data.csv",
				Key:  ".[0].firstname",
			},
			sourceInput:    "Tom",
			expectedResult: true,
		},
		{
			name: "Default successful workflow",
			spec: Spec{
				File:  "testdata/data.csv",
				Query: ".[*].firstname",
			},
			sourceInput:    "Tom",
			expectedResult: true,
		},
		{
			name: "Default successful workflow",
			spec: Spec{
				File:  "testdata/data.2.csv",
				Key:   ".[0].firstname",
				Comma: ';',
			},
			sourceInput:    "Tom",
			expectedResult: true,
		},
		{
			name: "Do not exist query workflow",
			spec: Spec{
				File:  "testdata/data.2.csv",
				Key:   ".[0].DoNotExist",
				Comma: ';',
			},
			sourceInput:      "Tom",
			expectedResult:   false,
			wantErr:          true,
			expectedErrorMsg: errors.New("could not find value for query \".[0].DoNotExist\" from file \"testdata/data.2.csv\""),
		},
		{
			name: "Changed workflow with Dasel v3",
			spec: Spec{
				File:   "testdata/data.csv",
				Key:    "$this[0].firstname",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			sourceInput:    "Tom",
			expectedResult: true,
		},
		{
			name: "Unchanged workflow with Dasel v3",
			spec: Spec{
				File:   "testdata/data.csv",
				Key:    "$this[0].firstname",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			sourceInput:    "John",
			expectedResult: false,
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {
			c, err := New(tt.spec)

			require.NoError(t, err)

			gotResult := result.Target{}
			err = c.Target(context.Background(), tt.sourceInput, nil, true, &gotResult)

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult.Changed)
		})
	}
}

// TestTargetDaselV3Write exercises the dasel v3 write path with a real (dryRun=false)
// write to disk. The dasel v3 engine stores modified values as *interface{}, so this
// verifies the CSV writer resolves them to plain strings (not pointer addresses) and
// leaves the rest of the document intact.
func TestTargetDaselV3Write(t *testing.T) {
	initialCSV := "firstname,surname,lastname\nJohn,,Smith\nAlexis,Alex,Remi\n"

	dir := t.TempDir()
	filePath := dir + "/data.csv"
	require.NoError(t, os.WriteFile(filePath, []byte(initialCSV), 0600))

	spec := Spec{
		File:   filePath,
		Key:    "$this[0].firstname",
		Value:  "Tom",
		Engine: strPtr(ENGINEDASEL_V3),
	}

	c, err := New(spec)
	require.NoError(t, err)

	gotResult := result.Target{}
	// dryRun=false so the file is actually written back to disk.
	err = c.Target(context.Background(), "Tom", nil, false, &gotResult)
	require.NoError(t, err)
	assert.True(t, gotResult.Changed)

	content, err := os.ReadFile(filePath)
	require.NoError(t, err)

	got := string(content)
	// The updated cell must be the plain value, not a pointer address (0x...).
	assert.Contains(t, got, "Tom,,Smith", "expected first row firstname updated to Tom, got:\n%s", got)
	assert.NotContains(t, got, "0x", "expected no pointer address leaked into the CSV, got:\n%s", got)
	// Untouched rows and headers survive.
	assert.Contains(t, got, "firstname,surname,lastname", "expected header preserved, got:\n%s", got)
	assert.Contains(t, got, "Alexis,Alex,Remi", "expected second row preserved, got:\n%s", got)
}

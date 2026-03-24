package json

import (
	"fmt"
	"os"
	"strings"
+	"context"
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
				File: "testdata/data.json",
				Key:  ".firstName",
			},
			sourceInput:    "Jack",
			expectedResult: false,
		},
		{
			name: "Default successful workflow using Dasel v2",
			spec: Spec{
				File:   "testdata/data.json",
				Key:    ".firstName",
				Engine: strPtr(ENGINEDASEL_V2),
			},
			sourceInput:    "Jack",
			expectedResult: false,
		},
		{
			name: "Default successful workflow",
			spec: Spec{
				File: "testdata/data.json",
				Key:  ".firstName",
			},
			sourceInput:    "Tom",
			expectedResult: true,
		},
		{
			name: "Default successful workflow",
			spec: Spec{
				File:  "testdata/data.json",
				Query: ".phoneNumbers.[*].type",
			},
			sourceInput:    "Tom",
			expectedResult: true,
		},
		{
			name: "Update first array item successful workflow",
			spec: Spec{
				File:   "testdata/data.json",
				Key:    ".phoneNumbers.first().type",
				Engine: strPtr(ENGINEDASEL_V2),
			},
			sourceInput:    "apartment",
			expectedResult: true,
		},
		{
			name: "Unchanged first array item successful workflow",
			spec: Spec{
				File:   "testdata/data.json",
				Key:    ".phoneNumbers.first().type",
				Engine: strPtr(ENGINEDASEL_V2),
			},
			sourceInput:    "home",
			expectedResult: false,
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {
			j, err := New(tt.spec)

			require.NoError(t, err)

			gotResult := result.Target{}
			err = j.Target(context.Background(), tt.sourceInput, nil, true, &gotResult)

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult.Changed)
		})
	}
}

// TestTargetPreservesSpecialCharacters verifies that HTML-special characters such as
// ">" are not escaped to their Unicode equivalents (e.g. \u003e) when the target
// writes back to disk. This is a regression test for the HTML-escaping bug.
func TestTargetPreservesSpecialCharacters(t *testing.T) {
	// Initial JSON contains a browsers key with ">0.2%" that must survive a write
	// cycle untouched, alongside a separate version key that the target will update.
	initialJSON := `{
  "browsers": ">0.2%",
  "version": "1.0.0"
}
`

	engines := []struct {
		name   string
		engine *string
	}{
		{name: "dasel/v1 (default)", engine: nil},
		{name: "dasel/v2", engine: strPtr(ENGINEDASEL_V2)},
	}

	for _, e := range engines {
		t.Run(fmt.Sprintf("engine=%s", e.name), func(t *testing.T) {
			dir := t.TempDir()
			filePath := dir + "/test.json"

			err := os.WriteFile(filePath, []byte(initialJSON), 0600)
			require.NoError(t, err)

			spec := Spec{
				File:   filePath,
				Key:    ".version",
				Engine: e.engine,
			}

			j, err := New(spec)
			require.NoError(t, err)

			gotResult := result.Target{}
			// dryRun=false so the file is actually written back to disk.
			err = j.Target("2.0.0", nil, false, &gotResult)
			require.NoError(t, err)
			assert.True(t, gotResult.Changed)

			content, err := os.ReadFile(filePath)
			require.NoError(t, err)

			// The literal ">" must be preserved; \u003e is the escaped form that
			// Go's encoding/json emits by default via its HTMLEscape behavior.
			assert.True(t, strings.Contains(string(content), ">0.2%"),
				"expected >0.2%% to be preserved in file, got:\n%s", string(content))
			assert.False(t, strings.Contains(string(content), `\u003e`),
				"expected > NOT to be HTML-escaped to \\u003e, got:\n%s", string(content))
		})
	}
}

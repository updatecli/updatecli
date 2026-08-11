package toml

import (
	"context"
	"errors"
	"os"
	"strings"
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
			name: "Deprecated multiple Test key do not exist",
			spec: Spec{
				File:     "testdata/data.toml",
				Key:      ".doNotExist.[*]",
				Value:    "",
				Multiple: true,
			},
			expectedResult:   false,
			sourceInput:      "M",
			wantErr:          true,
			expectedErrorMsg: errors.New("could not find multiple value for query \".doNotExist.[*]\" from file \"testdata/data.toml\""),
		},
		{
			name: "Test key do not exist",
			spec: Spec{
				File:  "testdata/data.toml",
				Query: ".doNotExist.[*]",
				Value: "",
			},
			expectedResult:   false,
			sourceInput:      "M",
			wantErr:          true,
			expectedErrorMsg: errors.New("could not find multiple value for query \".doNotExist.[*]\" from file \"testdata/data.toml\""),
		},
		{
			name: "Test key do not exist",
			spec: Spec{
				File:  "testdata/data.toml",
				Key:   ".doNotExist",
				Value: "",
			},
			expectedResult:   false,
			sourceInput:      "M",
			wantErr:          true,
			expectedErrorMsg: errors.New("could not find value for query \".doNotExist\" from file \"testdata/data.toml\""),
		},
		{
			name: "Default successful multiple update workflow",
			spec: Spec{
				File:  "testdata/data.toml",
				Query: ".employees.[*].role",
			},
			sourceInput:    "M",
			expectedResult: true,
		},
		{
			name: "Successful conditional multiple update workflow",
			spec: Spec{
				File:  "testdata/data.toml",
				Query: ".employees.(address=AU).role",
			},
			sourceInput:    "M",
			expectedResult: false,
		},
		{
			name: "Successful multiple map update workflow",
			spec: Spec{
				File:  "testdata/data.toml",
				Query: ".benefits.[0].country.(country=UK).name",
			},
			sourceInput:    "all",
			expectedResult: true,
		},
		{
			name: "Successful single update workflow",
			spec: Spec{
				File: "testdata/data.toml",
				Key:  ".owner.firstName",
			},
			sourceInput:    "Tom",
			expectedResult: true,
		},
		{
			name: "Successful no update workflow",
			spec: Spec{
				File: "testdata/data.toml",
				Key:  ".owner.firstName",
			},
			sourceInput:    "Jack",
			expectedResult: false,
		},
		{
			name: "Failing on non-existing key by default",
			spec: Spec{
				File: "testdata/data.toml",
				Key:  ".owner.age",
			},
			sourceInput:      "50",
			expectedResult:   false,
			wantErr:          true,
			expectedErrorMsg: errors.New("could not find value for query \".owner.age\" from file \"testdata/data.toml\""),
		},
		{
			name: "Successful update on non-existing key",
			spec: Spec{
				File:             "testdata/data.toml",
				Key:              ".owner.age",
				CreateMissingKey: true,
			},
			sourceInput:    "50",
			expectedResult: true,
		},
		{
			name: "Successful single update workflow with Dasel v3",
			spec: Spec{
				File:   "testdata/data.toml",
				Key:    "owner.firstName",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			sourceInput:    "Tom",
			expectedResult: true,
		},
		{
			name: "Successful no update workflow with Dasel v3",
			spec: Spec{
				File:   "testdata/data.toml",
				Key:    "owner.firstName",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			sourceInput:    "Jack",
			expectedResult: false,
		},
		{
			name: "Update nested array item with Dasel v3",
			spec: Spec{
				File:   "testdata/data.toml",
				Key:    "servers.beta.role",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			sourceInput:    "database",
			expectedResult: true,
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

// TestTargetDaselV3Write exercises the dasel v3 write path (PutV3 + WriteV3) with a
// real (dryRun=false) write to disk, ensuring the value is updated, the rest of the
// document (including a TOML datetime) survives the round-trip, and the output keeps
// a 2-space indentation for nested tables.
func TestTargetDaselV3Write(t *testing.T) {
	initialTOML := `title = "TOML Example"

[owner]
firstName = "Jack"
dob = 1979-05-27T07:32:00-08:00

[servers]

  [servers.beta]
  ip = "10.0.0.2"
  role = "backend"
`

	dir := t.TempDir()
	filePath := dir + "/data.toml"
	require.NoError(t, os.WriteFile(filePath, []byte(initialTOML), 0600))

	spec := Spec{
		File:   filePath,
		Key:    "owner.firstName",
		Engine: strPtr(ENGINEDASEL_V3),
	}

	j, err := New(spec)
	require.NoError(t, err)

	gotResult := result.Target{}
	// dryRun=false so the file is actually written back to disk.
	err = j.Target(context.Background(), "Tom", nil, false, &gotResult)
	require.NoError(t, err)
	assert.True(t, gotResult.Changed)

	content, err := os.ReadFile(filePath)
	require.NoError(t, err)

	got := string(content)
	assert.True(t, strings.Contains(got, `firstName = "Tom"`),
		"expected firstName to be updated to Tom, got:\n%s", got)
	// The datetime must survive the write cycle untouched.
	assert.True(t, strings.Contains(got, "1979-05-27T07:32:00-08:00"),
		"expected the TOML datetime to be preserved, got:\n%s", got)
	// Nested tables should be indented with 2 spaces.
	assert.True(t, strings.Contains(got, "  [servers.beta]"),
		"expected nested table to be indented with 2 spaces, got:\n%s", got)
}

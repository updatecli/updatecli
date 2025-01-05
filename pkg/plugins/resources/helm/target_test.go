package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestTarget(t *testing.T) {
	testData := []struct {
		name                      string
		spec                      Spec
		sourceInput               string
		expectedResult            bool
		expectedResultDescription string
		wantErr                   bool
	}{
		{
			name: "Success - No change",
			spec: Spec{
				Name:             "testdata",
				File:             "Chart.yaml",
				Key:              "$.dependencies[0].version",
				VersionIncrement: NOINCREMENT,
			},
			sourceInput:               "1.0.0",
			expectedResult:            false,
			expectedResultDescription: "no change detected:\n\t* key \"$.dependencies[0].version\" already set to \"1.0.0\", from file \"testdata/Chart.yaml\"",
		},
		{
			name: "Success - change with only App Version",
			spec: Spec{
				Name:             "testdata",
				File:             "Chart.yaml",
				Key:              "$.dependencies[0].version",
				VersionIncrement: NOINCREMENT,
				AppVersion:       true,
			},
			sourceInput:               "1.0.0",
			expectedResult:            true,
			expectedResultDescription: "no change detected:\n\t* key \"$.dependencies[0].version\" already set to \"1.0.0\", from file \"testdata/Chart.yaml\"\nchange detected:\n\t* key \"$.appVersion\" should be updated from \"0.1.0\" to \"1.0.0\", in file \"testdata/Chart.yaml\"",
		},
		{
			name: "Success - No change with Version Increment",
			spec: Spec{
				Name:             "testdata",
				File:             "Chart.yaml",
				Key:              "$.dependencies[0].version",
				VersionIncrement: MAJORVERSION,
			},
			sourceInput:               "1.0.0",
			expectedResult:            false,
			expectedResultDescription: "no change detected:\n\t* key \"$.dependencies[0].version\" already set to \"1.0.0\", from file \"testdata/Chart.yaml\"",
		},
		{
			name: "Success - No change with Version Increment and App Version",
			spec: Spec{
				Name:             "testdata",
				File:             "Chart.yaml",
				Key:              "$.dependencies[0].version",
				VersionIncrement: MAJORVERSION,
				AppVersion:       true,
			},
			sourceInput:               "1.0.0",
			expectedResult:            true,
			expectedResultDescription: "no change detected:\n\t* key \"$.dependencies[0].version\" already set to \"1.0.0\", from file \"testdata/Chart.yaml\"\nchange detected:\n\t* key \"$.appVersion\" should be updated from \"0.1.0\" to \"1.0.0\", in file \"testdata/Chart.yaml\"\nchange detected:\n\t* key \"$.version\" should be updated from \"0.3.0\" to \"1.0.0\", in file \"testdata/Chart.yaml\"",
		},
		{
			name: "Success - Expected change",
			spec: Spec{
				Name:             "testdata",
				File:             "Chart.yaml",
				Key:              "$.dependencies[0].version",
				VersionIncrement: NOINCREMENT,
			},
			sourceInput:               "1.1.0",
			expectedResult:            true,
			expectedResultDescription: "change detected:\n\t* key \"$.dependencies[0].version\" should be updated from \"1.0.0\" to \"1.1.0\", in file \"testdata/Chart.yaml\"",
		},
		{
			name: "Success - Expected change with Version Increment",
			spec: Spec{
				Name:             "testdata",
				File:             "Chart.yaml",
				Key:              "$.dependencies[0].version",
				VersionIncrement: MAJORVERSION,
			},
			sourceInput:               "1.1.0",
			expectedResult:            true,
			expectedResultDescription: "change detected:\n\t* key \"$.dependencies[0].version\" should be updated from \"1.0.0\" to \"1.1.0\", in file \"testdata/Chart.yaml\"\nchange detected:\n\t* key \"$.version\" should be updated from \"0.3.0\" to \"1.0.0\", in file \"testdata/Chart.yaml\"",
		},
		{
			name: "Success - Expected change with App Version",
			spec: Spec{
				Name:             "testdata",
				File:             "Chart.yaml",
				Key:              "$.dependencies[0].version",
				VersionIncrement: NOINCREMENT,
				AppVersion:       true,
			},
			sourceInput:               "1.1.0",
			expectedResult:            true,
			expectedResultDescription: "change detected:\n\t* key \"$.dependencies[0].version\" should be updated from \"1.0.0\" to \"1.1.0\", in file \"testdata/Chart.yaml\"\nchange detected:\n\t* key \"$.appVersion\" should be updated from \"0.1.0\" to \"1.1.0\", in file \"testdata/Chart.yaml\"",
		},
		{
			name: "Success - Expected change with Version Increment and App Version",
			spec: Spec{
				Name:             "testdata",
				File:             "Chart.yaml",
				Key:              "$.dependencies[0].version",
				VersionIncrement: MAJORVERSION,
				AppVersion:       true,
			},
			sourceInput:               "1.1.0",
			expectedResult:            true,
			expectedResultDescription: "change detected:\n\t* key \"$.dependencies[0].version\" should be updated from \"1.0.0\" to \"1.1.0\", in file \"testdata/Chart.yaml\"\nchange detected:\n\t* key \"$.appVersion\" should be updated from \"0.1.0\" to \"1.1.0\", in file \"testdata/Chart.yaml\"\nchange detected:\n\t* key \"$.version\" should be updated from \"0.3.0\" to \"1.0.0\", in file \"testdata/Chart.yaml\"",
		},
		{
			name: "Success - No change using Value",
			spec: Spec{
				Name:             "testdata",
				File:             "Chart.yaml",
				Key:              "$.dependencies[0].version",
				Value:            "1.0.0",
				VersionIncrement: NOINCREMENT,
			},
			expectedResult:            false,
			expectedResultDescription: "no change detected:\n\t* key \"$.dependencies[0].version\" already set to \"1.0.0\", from file \"testdata/Chart.yaml\"",
		},
		{
			name: "Success - Expected change using Value",
			spec: Spec{
				Name:             "testdata",
				File:             "Chart.yaml",
				Key:              "$.dependencies[0].version",
				VersionIncrement: NOINCREMENT,
				Value:            "1.1.0",
			},
			expectedResult:            true,
			expectedResultDescription: "change detected:\n\t* key \"$.dependencies[0].version\" should be updated from \"1.0.0\" to \"1.1.0\", in file \"testdata/Chart.yaml\"",
		},
		{
			name: "Success - Expected change using Value with Version Increment",
			spec: Spec{
				Name:             "testdata",
				File:             "Chart.yaml",
				Key:              "$.dependencies[0].version",
				VersionIncrement: MAJORVERSION,
				Value:            "1.1.0",
			},
			expectedResult:            true,
			expectedResultDescription: "change detected:\n\t* key \"$.dependencies[0].version\" should be updated from \"1.0.0\" to \"1.1.0\", in file \"testdata/Chart.yaml\"\nchange detected:\n\t* key \"$.version\" should be updated from \"0.3.0\" to \"1.0.0\", in file \"testdata/Chart.yaml\"",
		},
		{
			name: "Success - Expected change using Value with Version Increment and App Version",
			spec: Spec{
				Name:             "testdata",
				File:             "Chart.yaml",
				Key:              "$.dependencies[0].version",
				VersionIncrement: MAJORVERSION,
				AppVersion:       true,
				Value:            "1.1.0",
			},
			expectedResult:            true,
			expectedResultDescription: "change detected:\n\t* key \"$.dependencies[0].version\" should be updated from \"1.0.0\" to \"1.1.0\", in file \"testdata/Chart.yaml\"\nchange detected:\n\t* key \"$.appVersion\" should be updated from \"0.1.0\" to \"1.1.0\", in file \"testdata/Chart.yaml\"\nchange detected:\n\t* key \"$.version\" should be updated from \"0.3.0\" to \"1.0.0\", in file \"testdata/Chart.yaml\"",
		},
		{
			name: "Success - Expected change with Version Increment to values.yaml",
			spec: Spec{
				Name:             "testdata",
				File:             "values.yaml",
				Key:              "$.version",
				VersionIncrement: MAJORVERSION,
			},
			sourceInput:               "1.1.0",
			expectedResult:            true,
			expectedResultDescription: "change detected:\n\t* key \"$.version\" should be updated from \"1.0.0\" to \"1.1.0\", in file \"testdata/values.yaml\"\nchange detected:\n\t* key \"$.version\" should be updated from \"0.3.0\" to \"1.0.0\", in file \"testdata/Chart.yaml\"",
		},
		{
			name: "Success - Child chart patch update Version Increment to Chart.yaml",
			spec: Spec{
				Name:             "testdata",
				File:             "Chart.yaml",
				Key:              "$.dependencies[0].version",
				VersionIncrement: AUTO,
			},
			sourceInput:               "1.0.1",
			expectedResult:            true,
			expectedResultDescription: "change detected:\n\t* key \"$.dependencies[0].version\" should be updated from \"1.0.0\" to \"1.0.1\", in file \"testdata/Chart.yaml\"\nchange detected:\n\t* key \"$.version\" should be updated from \"0.3.0\" to \"0.3.1\", in file \"testdata/Chart.yaml\"",
		},
		{
			name: "Success - Auto chart minor update Version Increment to Chart.yaml",
			spec: Spec{
				Name:             "testdata",
				File:             "Chart.yaml",
				Key:              "$.dependencies[0].version",
				VersionIncrement: AUTO,
			},
			sourceInput:               "1.1.0",
			expectedResult:            true,
			expectedResultDescription: "change detected:\n\t* key \"$.dependencies[0].version\" should be updated from \"1.0.0\" to \"1.1.0\", in file \"testdata/Chart.yaml\"\nchange detected:\n\t* key \"$.version\" should be updated from \"0.3.0\" to \"0.4.0\", in file \"testdata/Chart.yaml\"",
		},
		{
			name: "Success - Auto chart major update Version Increment to Chart.yaml",
			spec: Spec{
				Name:             "testdata",
				File:             "Chart.yaml",
				Key:              "$.dependencies[0].version",
				VersionIncrement: AUTO,
			},
			sourceInput:               "2.0.0",
			expectedResult:            true,
			expectedResultDescription: "change detected:\n\t* key \"$.dependencies[0].version\" should be updated from \"1.0.0\" to \"2.0.0\", in file \"testdata/Chart.yaml\"\nchange detected:\n\t* key \"$.version\" should be updated from \"0.3.0\" to \"1.0.0\", in file \"testdata/Chart.yaml\"",
		},
		{
			name: "Success - Auto chart major update with tilde Version Increment to Chart.yaml",
			spec: Spec{
				Name:             "testdata",
				File:             "Chart.yaml",
				Key:              "$.dependencies[0].version",
				VersionIncrement: AUTO,
			},
			sourceInput:               "~2.0.0",
			expectedResult:            true,
			expectedResultDescription: "change detected:\n\t* key \"$.dependencies[0].version\" should be updated from \"1.0.0\" to \"~2.0.0\", in file \"testdata/Chart.yaml\"\nchange detected:\n\t* key \"$.version\" should be updated from \"0.3.0\" to \"1.0.0\", in file \"testdata/Chart.yaml\"",
		},
		{
			name: "Success - Auto chart minor update with >= Version Increment to Chart.yaml",
			spec: Spec{
				Name:             "testdata",
				File:             "Chart.yaml",
				Key:              "$.dependencies[0].version",
				VersionIncrement: AUTO,
			},
			sourceInput:               `">=1.1.0"`,
			expectedResult:            true,
			expectedResultDescription: "change detected:\n\t* key \"$.dependencies[0].version\" should be updated from \"1.0.0\" to \"\\\">=1.1.0\\\"\", in file \"testdata/Chart.yaml\"\nchange detected:\n\t* key \"$.version\" should be updated from \"0.3.0\" to \"0.4.0\", in file \"testdata/Chart.yaml\"",
		},
		{
			name: "Success - Auto chart major and patch update Version Increment to Chart.yaml",
			spec: Spec{
				Name:             "testdata",
				File:             "Chart.yaml",
				Key:              "$.dependencies[0].version",
				VersionIncrement: AUTO,
			},
			sourceInput:               "2.0.5",
			expectedResult:            true,
			expectedResultDescription: "change detected:\n\t* key \"$.dependencies[0].version\" should be updated from \"1.0.0\" to \"2.0.5\", in file \"testdata/Chart.yaml\"\nchange detected:\n\t* key \"$.version\" should be updated from \"0.3.0\" to \"1.0.0\", in file \"testdata/Chart.yaml\"",
		},
		{
			name: "Success - Auto chart minor update Version Increment to Chart.yaml due to invalid semver target version",
			spec: Spec{
				Name:             "testdata",
				File:             "Chart.yaml",
				Key:              "$.dependencies[0].version",
				VersionIncrement: AUTO,
			},
			sourceInput:               "invalid.semver",
			expectedResult:            true,
			expectedResultDescription: "change detected:\n\t* key \"$.dependencies[0].version\" should be updated from \"1.0.0\" to \"invalid.semver\", in file \"testdata/Chart.yaml\"\nchange detected:\n\t* key \"$.version\" should be updated from \"0.3.0\" to \"0.4.0\", in file \"testdata/Chart.yaml\"",
		},
		{
			name: "Success - Auto chart major downgrade Version Increment to Chart.yaml",
			spec: Spec{
				Name:             "testdata",
				File:             "values.yaml",
				Key:              "$.otherVersion",
				VersionIncrement: AUTO,
			},
			sourceInput:               "0.2.3",
			expectedResult:            true,
			expectedResultDescription: "change detected:\n\t* key \"$.otherVersion\" should be updated from \"1.2.3\" to \"0.2.3\", in file \"testdata/values.yaml\"\nchange detected:\n\t* key \"$.version\" should be updated from \"0.3.0\" to \"1.0.0\", in file \"testdata/Chart.yaml\"",
		},
		{
			name: "Success - Auto chart minor downgrade Version Increment to Chart.yaml",
			spec: Spec{
				Name:             "testdata",
				File:             "values.yaml",
				Key:              "$.otherVersion",
				VersionIncrement: AUTO,
			},
			sourceInput:               "1.1.3",
			expectedResult:            true,
			expectedResultDescription: "change detected:\n\t* key \"$.otherVersion\" should be updated from \"1.2.3\" to \"1.1.3\", in file \"testdata/values.yaml\"\nchange detected:\n\t* key \"$.version\" should be updated from \"0.3.0\" to \"0.4.0\", in file \"testdata/Chart.yaml\"",
		},
		{
			name: "Success - Auto chart patch downgrade Version Increment to Chart.yaml",
			spec: Spec{
				Name:             "testdata",
				File:             "values.yaml",
				Key:              "$.otherVersion",
				VersionIncrement: AUTO,
			},
			sourceInput:               "1.2.2",
			expectedResult:            true,
			expectedResultDescription: "change detected:\n\t* key \"$.otherVersion\" should be updated from \"1.2.3\" to \"1.2.2\", in file \"testdata/values.yaml\"\nchange detected:\n\t* key \"$.version\" should be updated from \"0.3.0\" to \"0.3.1\", in file \"testdata/Chart.yaml\"",
		},
		{
			name: "Success - Auto chart semver metadata only change Version Increment to Chart.yaml",
			spec: Spec{
				Name:             "testdata",
				File:             "values.yaml",
				Key:              "$.otherVersion",
				VersionIncrement: AUTO,
			},
			sourceInput:               "1.2.3-rc1",
			expectedResult:            true,
			expectedResultDescription: "change detected:\n\t* key \"$.otherVersion\" should be updated from \"1.2.3\" to \"1.2.3-rc1\", in file \"testdata/values.yaml\"\nchange detected:\n\t* key \"$.version\" should be updated from \"0.3.0\" to \"0.4.0\", in file \"testdata/Chart.yaml\"",
		},
		{
			name: "Failure - File does not exists",
			spec: Spec{
				Name:             "testdata",
				File:             "doNotExist.yaml",
				Key:              "$.dependencies[0].version",
				VersionIncrement: NOINCREMENT,
			},
			expectedResult: false,
			wantErr:        true,
		},
		{
			name: "Failure - Path does not exists",
			spec: Spec{
				Name:             "testdata",
				File:             "Chart.yaml",
				Key:              "$.dependencies[1].version",
				VersionIncrement: NOINCREMENT,
			},
			expectedResult: false,
			wantErr:        true,
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			j, err := New(tt.spec)

			require.NoError(t, err)

			gotResult := result.Target{}
			err = j.Target(tt.sourceInput, nil, true, &gotResult)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult.Changed)
			assert.Equal(t, tt.expectedResultDescription, gotResult.Description)
		})
	}
}

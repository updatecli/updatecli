package xml

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

type TargetDataset []TargetData

type TargetData struct {
	name         string
	data         XML
	wantResult   bool
	wantErrorMsg error
	wantErr      bool
}

var (
	targetDataset = TargetDataset{
		{
			name: "Test 1",
			data: XML{
				spec: Spec{
					File:  "testdata/data_0.xml",
					Path:  "/name/firstname",
					Value: "Bob",
				},
			},
			wantResult: true,
		},
		{
			name: "Test 2",
			data: XML{
				spec: Spec{
					File:  "testdata/data_0.xml",
					Path:  "/name/firstname",
					Value: "John",
				},
			},
			wantResult: false,
		},
		{
			name: "Test 3",
			data: XML{
				spec: Spec{
					File:  "testdata/data_2.xml",
					Path:  "/name/firstname",
					Value: "Bob",
				},
			},
			wantResult: true,
		},
		{
			name: "Test 4",
			data: XML{
				spec: Spec{
					File:  "testdata/doNotExist.xml",
					Path:  "/name/firstname",
					Value: "Alice",
				},
			},
			wantResult:   false,
			wantErrorMsg: errors.New("open testdata/doNotExist.xml: no such file or directory"),
			wantErr:      true,
		},
		{
			data: XML{
				spec: Spec{
					File:  "testdata/data_2.xml",
					Path:  "/name/donotexist",
					Value: "Bob",
				},
			},
			wantResult:   false,
			wantErr:      true,
			wantErrorMsg: errors.New("✗ nothing found at path \"/name/donotexist\" from file \"testdata/data_2.xml\""),
		},
		{
			data: XML{
				spec: Spec{
					File:  "testdata/data_2.xml",
					Path:  "/name/firstname",
					Value: "John",
				},
			},
			wantResult: false,
		},
	}
)

func TestTarget(t *testing.T) {

	for _, tt := range targetDataset {

		t.Run(tt.name, func(t *testing.T) {

			gotResult, gotErr := tt.data.Target("", true)

			if tt.wantErr {
				require.Error(t, gotErr)
			} else {
				require.NoError(t, gotErr)
			}

			assert.Equal(t, tt.wantResult, gotResult)

		})
	}

}

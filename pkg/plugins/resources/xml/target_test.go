package xml

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
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
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {
			x, err := New(tt.spec)

			require.NoError(t, err)

			gotResult := result.Target{}
			err = x.Target("", nil, true, &gotResult)

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult.Changed)
		})
	}

}

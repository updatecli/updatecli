package xml

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestSource(t *testing.T) {

	testData := []struct {
		name             string
		spec             Spec
		expectedResult   string
		wantErr          bool
		expectedErrorMsg string
	}{
		{
			name: "scenario 1",
			spec: Spec{
				File: "testdata/data_0.xml",
				Path: "//name/firstname",
			},
			expectedResult: "John",
		},
		{
			name: "scenario 1.1 - http",
			spec: Spec{
				File: "https://raw.githubusercontent.com/updatecli/updatecli/main/pkg/plugins/resources/xml/testdata/data_0.xml",
				Path: "//name/firstname",
			},
			expectedResult: "John",
		},
		{
			name: "scenario 2",
			spec: Spec{
				File: "testdata/data_1.xml",
				Path: "doNotExist",
			},
			expectedResult:   "",
			wantErr:          true,
			expectedErrorMsg: "cannot find value for path \"doNotExist\" from file \"testdata/data_1.xml\"",
		},
		{
			name: "scenario 3",
			spec: Spec{
				File: "testdata/data_1.xml",
				Path: "//breakfast_menu/food[0]/name",
			},
			expectedResult: "Belgian Waffles",
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {
			x, err := New(tt.spec)

			require.NoError(t, err)

			gotResult := result.Source{}
			err = x.Source("", &gotResult)

			switch tt.wantErr {
			case true:
				require.ErrorContains(t, err, tt.expectedErrorMsg)
			case false:
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult.Information)
		})
	}
}

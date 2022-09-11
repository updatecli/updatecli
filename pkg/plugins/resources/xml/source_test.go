package xml

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestSource(t *testing.T) {

	testData := []struct {
		name           string
		spec           Spec
		expectedResult string
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
			expectedResult: "",
		},
		{
			name: "scenario 3",
			spec: Spec{
				File: "testdata/data_1.xml",
				Path: "//breakfast_menu[0]/name",
			},
			expectedResult: "",
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {
			x, err := New(tt.spec)

			require.NoError(t, err)

			gotResult, err := x.Source("")

			require.NoError(t, err)

			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}

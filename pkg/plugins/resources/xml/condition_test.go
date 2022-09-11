package xml

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestCondition(t *testing.T) {

	testData := []struct {
		name             string
		spec             Spec
		expectedResult   bool
		expectedErrorMsg error
		wantErr          bool
	}{
		{
			spec: Spec{
				File: "testdata/data_0.xml",
				Path: "/name/firstname",
			},
			expectedResult: false,
		},
		{
			spec: Spec{
				File:  "testdata/data_0.xml",
				Path:  "/name/firstname",
				Value: "John",
			},
			expectedResult: true,
		},
		{
			spec: Spec{
				File:  "https://raw.githubusercontent.com/updatecli/updatecli/main/pkg/plugins/resources/xml/testdata/data_0.xml",
				Path:  "/name/firstname",
				Value: "John",
			},
			expectedResult: true,
		},
		{
			spec: Spec{
				File:  "testdata/data_0.xml",
				Path:  "/name/firstname",
				Value: "wrongValue",
			},
			expectedResult: false,
		},
		{
			spec: Spec{
				File: "testdata/data_0.xml",
				Path: ".name.donotExit",
			},
			expectedResult: false,
		},
		{
			spec: Spec{
				File: "testdata/data_1.xml",
				Path: "/breakfast_menu/food[0]/name",
			},
			expectedResult: false,
		},
		{
			spec: Spec{
				File:  "testdata/data_1.xml",
				Path:  "/breakfast_menu/food[0]/name",
				Value: "Belgian Waffles",
			},
			expectedResult: true,
		},
		{
			spec: Spec{
				File:  "testdata/data_1.xml",
				Path:  "/breakfast_menu.food[0]/name",
				Value: "wrongValue",
			},
			expectedResult: false,
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {
			x, err := New(tt.spec)

			require.NoError(t, err)

			gotResult, err := x.Condition("")

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}

}

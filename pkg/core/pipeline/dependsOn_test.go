package pipeline

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseDependsOn(t *testing.T) {
	testdata := []struct {
		dependsOn               string
		expectedKey             string
		expectedBooleanOperator string
	}{
		{
			dependsOn:               "example",
			expectedKey:             "example",
			expectedBooleanOperator: "and",
		},
		{
			dependsOn:               "example:or",
			expectedKey:             "example",
			expectedBooleanOperator: "or",
		},
		{
			dependsOn:               "example:or:or",
			expectedKey:             "example:or",
			expectedBooleanOperator: "or",
		},
		{
			dependsOn:               "example:or:or:or",
			expectedKey:             "example:or:or",
			expectedBooleanOperator: "or",
		},
		{
			dependsOn:               "",
			expectedKey:             "",
			expectedBooleanOperator: "",
		},
	}

	for _, tt := range testdata {
		t.Run(tt.dependsOn, func(t *testing.T) {
			gotKey, gotBooleanOperator := parseDependsOnValue(tt.dependsOn)

			require.Equal(t, tt.expectedKey, gotKey)
			require.Equal(t, tt.expectedBooleanOperator, gotBooleanOperator)
		})
	}
}

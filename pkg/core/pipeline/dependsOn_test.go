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
		expectedCategory        string
	}{
		{
			dependsOn:               "example",
			expectedKey:             "example",
			expectedBooleanOperator: andBooleanOperator,
			expectedCategory:        "",
		},
		{
			dependsOn:               "example:or",
			expectedKey:             "example",
			expectedBooleanOperator: orBooleanOperator,
			expectedCategory:        "",
		},
		{
			dependsOn:               "example:or:or",
			expectedKey:             "example:or",
			expectedBooleanOperator: orBooleanOperator,
			expectedCategory:        "",
		},
		{
			dependsOn:               "example:or:or:or",
			expectedKey:             "example:or:or",
			expectedBooleanOperator: orBooleanOperator,
			expectedCategory:        "",
		},
		{
			dependsOn:               "",
			expectedKey:             "",
			expectedBooleanOperator: "",
			expectedCategory:        "",
		},
		{
			dependsOn:               "#example:or:or:or",
			expectedKey:             "example:or:or",
			expectedBooleanOperator: orBooleanOperator,
			expectedCategory:        "",
		},
		{
			dependsOn:               "source#example:or:or:or",
			expectedKey:             "example:or:or",
			expectedBooleanOperator: orBooleanOperator,
			expectedCategory:        "source",
		},
		{
			dependsOn:               "category#source#example:or:or:or",
			expectedKey:             "source#example:or:or",
			expectedBooleanOperator: orBooleanOperator,
			expectedCategory:        "category",
		},
	}

	for _, tt := range testdata {
		t.Run(tt.dependsOn, func(t *testing.T) {
			gotKey, gotBooleanOperator, _ := parseDependsOnValue(tt.dependsOn)

			require.Equal(t, tt.expectedKey, gotKey)
			require.Equal(t, tt.expectedBooleanOperator, gotBooleanOperator)
		})
	}
}

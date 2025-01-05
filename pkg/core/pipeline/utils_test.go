package pipeline

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractDepsFromTemplate(t *testing.T) {
	testdata := []struct {
		Name           string
		Template       string
		ExpectedResult []string
		ExpectedErr    string
	}{{
		Name: "Scenario 1",
		Template: `
{{ source "sourceId1" }}
		{{ condition "conditionid1" }}
		{{ target "targetid1" }}
		{{ source "sourceId2" }}
		{{ target "targetid2" }}
		{{ condition "conditionid2" }}
        `,
		ExpectedResult: []string{
			"source#sourceId1", "source#sourceId2",
			"condition#conditionid1", "condition#conditionid2",
			"target#targetid1", "target#targetid2",
		},
	},
	}

	for _, data := range testdata {

		t.Run(data.Name, func(t *testing.T) {
			got, err := ExtractDepsFromTemplate(data.Template)
			if data.ExpectedErr != "" {
				require.EqualError(t, err, data.ExpectedErr)
			} else {
				require.NoError(t, err)
			}
			sort.Strings(data.ExpectedResult)
			sort.Strings(got)
			require.Equal(t, data.ExpectedResult, got)
		})
	}
}

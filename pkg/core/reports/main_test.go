package reports

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestSortReports(t *testing.T) {
	reports := Reports{
		Report{Result: result.SUCCESS},
		Report{Result: result.ATTENTION},
		Report{Result: result.SKIPPED},
		Report{Result: result.FAILURE},
		Report{Result: result.SUCCESS},
		Report{Result: result.SUCCESS},
	}

	sorted := Reports{
		Report{Result: result.SUCCESS},
		Report{Result: result.SUCCESS},
		Report{Result: result.SUCCESS},
		Report{Result: result.SKIPPED},
		Report{Result: result.ATTENTION},
		Report{Result: result.FAILURE},
	}

	getResults := func(result Reports) []string {
		var r []string
		for _, report := range result {
			r = append(r, report.Result)
		}
		return r
	}

	reports.Sort()

	reportsResults := getResults(reports)
	sortedResults := getResults(sorted)

	assert.Equal(t, sortedResults, reportsResults, "Reports are not sorted")
}

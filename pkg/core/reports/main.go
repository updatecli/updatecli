package reports

import (
	"slices"
	"sort"

	"github.com/updatecli/updatecli/pkg/core/log"

	"github.com/updatecli/updatecli/pkg/core/result"
)

// Reports contains a list of report
type Reports []Report

// Show logs a short summary of every pipeline report
func (r *Reports) Show() error {
	log.Title("Summary")
	// Banners own the spacing around the summary
	log.Infof("%s", r.render()+"\n")

	return nil
}

// Sort reports by result in the following order
// SUCCESS > SKIPPED > ATTENTION > FAILURE > UNKNOWN
func (r *Reports) Sort() {
	reports := *r

	sort.SliceStable(reports, func(i, j int) bool {
		resultToInteger := func(state string) int {
			switch state {
			case result.SUCCESS:
				return 0
			case result.SKIPPED:
				return 1
			case result.ATTENTION:
				return 2
			case result.FAILURE:
				return 3
			default:
				return 4
			}
		}

		return resultToInteger(reports[i].Result) < resultToInteger(reports[j].Result)
	})

}

// Summary display a summary of
func (r *Reports) Summary() (successCounter, changedCounter, failedCounter, skippedCounter int, actionLinks []string) {

	reports := *r

	for _, report := range reports {
		switch report.Result {

		case result.SUCCESS:
			successCounter++

		case result.FAILURE:
			failedCounter++

		case result.SKIPPED:
			skippedCounter++

		case result.ATTENTION:
			changedCounter++

		default:
			log.Infof("Unknown report result %q with named %q.", report.Result, report.Name)
		}

		for _, action := range report.Actions {
			if !slices.Contains(actionLinks, action.Link) && action.Link != "" {
				actionLinks = append(actionLinks, action.Link)
			}
		}
	}

	return successCounter, changedCounter, failedCounter, skippedCounter, actionLinks
}

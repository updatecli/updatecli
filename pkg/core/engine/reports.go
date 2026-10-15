package engine

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/log"
)

// showReports display the reports
// and return an error if at least one pipeline failed
func (e *Engine) showReports() error {
	e.Reports.Sort()

	err := e.Reports.Show()
	if err != nil {
		return err
	}

	totalSuccessPipeline, totalChangedAppliedPipeline, totalFailedPipeline, totalSkippedPipeline, actionLinks := e.Reports.Summary()

	totalPipeline := totalSuccessPipeline + totalChangedAppliedPipeline + totalFailedPipeline + totalSkippedPipeline

	log.Heading("Run Summary")
	log.Infof("Pipeline(s) run:")
	log.Infof("  * Changed:\t%d", totalChangedAppliedPipeline)
	log.Infof("  * Failed:\t%d", totalFailedPipeline)
	log.Infof("  * Skipped:\t%d", totalSkippedPipeline)
	log.Infof("  * Succeeded:\t%d", totalSuccessPipeline)
	log.Infof("  * Total:\t%d\n\n", totalPipeline)

	switch len(actionLinks) {
	case 0:
		//
	case 1:
		log.Infof("One action to follow up:")
		for i := range actionLinks {
			log.Infof("  * %s", actionLinks[i])
		}
	default:
		log.Infof("%d action(s) to follow up:", len(actionLinks))
		for i := range actionLinks {
			log.Infof("  * %s", actionLinks[i])
		}
	}

	// Exit on error if at least one pipeline failed
	if totalFailedPipeline > 0 {
		return fmt.Errorf("%d over %d pipeline failed", totalFailedPipeline, totalPipeline)
	}

	return nil
}

package engine

import (
	"fmt"

	"github.com/sirupsen/logrus"
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

	logrus.Infof("Run Summary")
	logrus.Infof("===========\n")
	logrus.Infof("Pipeline(s) run:")
	logrus.Infof("  * Changed:\t%d", totalChangedAppliedPipeline)
	logrus.Infof("  * Failed:\t%d", totalFailedPipeline)
	logrus.Infof("  * Skipped:\t%d", totalSkippedPipeline)
	logrus.Infof("  * Succeeded:\t%d", totalSuccessPipeline)
	logrus.Infof("  * Total:\t%d\n\n", totalPipeline)

	switch len(actionLinks) {
	case 0:
		//
	case 1:
		logrus.Infof("One action to follow up:")
		for i := range actionLinks {
			logrus.Infof("  * %s", actionLinks[i])
		}
	default:
		logrus.Infof("%d action(s) to follow up:", len(actionLinks))
		for i := range actionLinks {
			logrus.Infof("  * %s", actionLinks[i])
		}
	}

	// Exit on error if at least one pipeline failed
	if totalFailedPipeline > 0 {
		return fmt.Errorf("%d over %d pipeline failed", totalFailedPipeline, totalPipeline)
	}

	return nil
}

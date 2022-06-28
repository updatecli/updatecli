package engine

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// ManifestUpgrade load Updatecli Manifest to update them then written them back on disk
func (e *Engine) ManifestUpgrade(saveToDisk bool) (err error) {
	logrus.Infof("\n\n%s\n", strings.Repeat("+", len("Manifest Upgrade")+4))
	logrus.Infof("+ %s +\n", strings.ToTitle("Manifest Upgrade"))
	logrus.Infof("%s\n\n", strings.Repeat("+", len("Manifest Upgrade")+4))

	err = e.LoadConfigurations()

	if len(e.Pipelines) == 0 {
		logrus.Errorln(err)
		return fmt.Errorf("no valid pipeline found")
	}

	// Don't exit if we identify at least one valid pipeline configuration
	if err != nil {
		logrus.Errorln(err)
		logrus.Infof("\n%d pipeline(s) successfully loaded\n", len(e.Pipelines))
	}

	for _, pipeline := range e.Pipelines {

		isManifestDifferentThanOnDisk, err := pipeline.Config.IsManifestDifferentThanOnDisk()

		if err != nil {
			pipeline.Report.Result = result.FAILURE
			e.Reports = append(e.Reports, pipeline.Report)
			logrus.Errorf("skipping manifest upgrade because of %q", err)
			continue
		}

		if saveToDisk && isManifestDifferentThanOnDisk {
			err = pipeline.Config.SaveOnDisk()
			if err != nil {
				pipeline.Report.Result = result.FAILURE
				e.Reports = append(e.Reports, pipeline.Report)
				logrus.Errorf("failed writing manifest change back to disk because of %q", err)
				continue
			}
		}

		switch isManifestDifferentThanOnDisk {
		case true:
			pipeline.Report.Result = result.ATTENTION
		default:
			pipeline.Report.Result = result.SUCCESS
		}

		e.Reports = append(e.Reports, pipeline.Report)
	}

	totalSuccessUpgrade, totalChangedAppliedUpgrade, totalFailedUpgrade, totalSkippedUpgrade := e.Reports.Summary()

	totalUpgrade := totalSuccessUpgrade + totalChangedAppliedUpgrade + totalFailedUpgrade + totalSkippedUpgrade

	logrus.Infof("\n\n\n")
	logrus.Infof("Summary")
	logrus.Infof("===========\n")
	logrus.Infof("Manifest(s) upgrade:")
	logrus.Infof("  * Changed:\t%d", totalChangedAppliedUpgrade)
	logrus.Infof("  * Failed:\t%d", totalFailedUpgrade)
	logrus.Infof("  * Skipped:\t%d", totalSkippedUpgrade)
	logrus.Infof("  * Succeeded:\t%d", totalSuccessUpgrade)
	logrus.Infof("  * Total:\t%d", totalUpgrade)

	// Exit on error if at least one upgrade failed
	if totalFailedUpgrade > 0 {
		return fmt.Errorf("%d over %d manifest upgrade failed", totalFailedUpgrade, totalUpgrade)
	}

	return err
}

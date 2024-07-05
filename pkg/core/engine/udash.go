package engine

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/cmdoptions"
	"github.com/updatecli/updatecli/pkg/core/udash"
)

// publishToUdash publish pipeline reports to the Udash service.
// This service is still experimental and should be used with caution.
// More information on https://github.com/updatecli/udash
func (e *Engine) publishToUdash() error {

	errs := []string{}

	if !cmdoptions.Experimental {
		return nil
	}

	logrus.Infof("\n\n%s\n", strings.ToTitle("Udash - Experimental"))
	logrus.Infof("%s\n\n", strings.Repeat("=", len("Udash - Experimental")+1))

	udashConfigFile, found := udash.IsConfigFile()
	if !found {
		logrus.Infof("Skipping as no Udash configuration file found at %s", udashConfigFile)
		return nil
	}

	for id := range e.Pipelines {
		pipeline := e.Pipelines[id]
		if err := udash.Publish(&pipeline.Report); err != nil &&
			!errors.Is(err, udash.ErrNoUdashAPIURL) {
			errs = append(errs, pipeline.Name+err.Error())
		}
		if pipeline.Report.ReportURL != "" {
			logrus.Infof("%s:\n\t=> %q", pipeline.Name, pipeline.Report.ReportURL)
		}
		e.Pipelines[id] = pipeline
	}

	if len(errs) > 0 {
		return fmt.Errorf(
			"errors occurred while publishing to Udash:\n\t* %s",
			strings.Join(errs, "\n\t* "),
		)
	}

	return nil
}

package engine

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/cmdoptions"
	"github.com/updatecli/updatecli/pkg/core/udash"
)

// ErrUdashPublish is returned when at least one pipeline report could not be
// published to Udash. Publishing is a side effect, so this error never changes
// a pipeline result.
var ErrUdashPublish = errors.New("publishing reports to Udash failed")

// publishToUdash publish pipeline reports to the Udash service.
// This service is still experimental and should be used with caution.
// More information on https://github.com/updatecli/udash
func (e *Engine) publishToUdash() error {

	if !cmdoptions.Experimental {
		if udash.IsConfigured() {
			logrus.Warningf("A Udash endpoint is configured but publishing reports requires the flag --experimental, skipping. Use --disable-udash-report to silence this warning.")
		}
		return nil
	}

	logrus.Infof("\n\n%s\n", strings.ToTitle("Udash - Experimental"))
	logrus.Infof("%s\n\n", strings.Repeat("=", len("Udash - Experimental")+1))

	errs := []error{}

	for id := range e.Pipelines {
		pipeline := e.Pipelines[id]
		err := udash.Publish(&pipeline.Report)
		if err != nil {
			if errors.Is(err, udash.ErrNoUdashAPIURL) {
				logrus.Infof("no Udash endpoint detected, skipping")
				break
			}
			errs = append(errs, fmt.Errorf("pipeline %q: %w", pipeline.Name, err))
		}
		if pipeline.Report.ReportURL != "" {
			logrus.Infof("%s:\n\t=> %q", pipeline.Name, pipeline.Report.ReportURL)
		}
		e.Pipelines[id] = pipeline
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w: %w", ErrUdashPublish, errors.Join(errs...))
	}

	return nil
}

// PublishErr returns the error met while publishing reports to Udash during the
// last run, if any. It is kept out of the error returned by Run.
func (e *Engine) PublishErr() error {
	return e.publishErr
}

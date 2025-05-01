package engine

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/cmdoptions"
)

// exportReportToYAML is a function that exports the report of the pipeline to a specified format and location.
func (e *Engine) exportReportToYAML(filenameTimestamp bool) error {
	errs := []string{}

	if !cmdoptions.Experimental {
		return nil
	}

	logrus.Infof("\n\n%s\n", strings.ToTitle("Report - Experimental"))
	logrus.Infof("%s\n\n", strings.Repeat("=", len("Report - Experimental")+1))

	for id := range e.Pipelines {
		pipeline := e.Pipelines[id]
		reportFilepath, err := pipeline.Report.ExportToYAML("", filenameTimestamp)
		if err != nil {
			errs = append(errs, pipeline.Name+err.Error())
		}
		if reportFilepath != "" {
			logrus.Infof("%s:\n\t=> %q", pipeline.Name, reportFilepath)
		}
		e.Pipelines[id] = pipeline
	}

	if len(errs) > 0 {
		return fmt.Errorf(
			"errors occurred while exporting report:\n\t* %s",
			strings.Join(errs, "\n\t* "),
		)
	}
	return nil
}

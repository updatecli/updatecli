package engine

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// exportReportToYAML is a function that exports the report of the pipeline to a specified format and location.
func (e *Engine) exportReportToYAML() error {
	errs := []string{}

	logrus.Infof("\n\n%s\n", strings.ToTitle("Report"))
	logrus.Infof("%s\n\n", strings.Repeat("=", len("Report")+1))

	for id := range e.Pipelines {
		pipeline := e.Pipelines[id]
		reportFilepath, err := pipeline.Report.ExportToYAML("")
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

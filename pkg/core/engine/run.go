package engine

import (
	"github.com/sirupsen/logrus"
)

// Run runs the full process
func (e *Engine) Run() (err error) {

	PrintTitle("Pipeline")

	for _, pipeline := range e.Pipelines {

		err := pipeline.Run()

		e.Reports = append(e.Reports, pipeline.Report)

		if err != nil {
			logrus.Printf("Pipeline %q failed\n", pipeline.Name)
			logrus.Printf("Skipping due to:\n\t%s\n", err)
			continue
		}
	}

	if err = e.runActions(); err != nil {
		logrus.Errorf("running actions:\n%s", err)
	}

	if err = e.publishToUdash(); err != nil {
		logrus.Errorf("publishing to Udash:\n%s", err)
	}

	if err = e.showReports(); err != nil {
		return err
	}

	return nil
}

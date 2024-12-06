package engine

import (
	"github.com/sirupsen/logrus"
)

// Run runs the full process
func (e *Engine) Run() (err error) {

	PrintTitle("Pipeline")

	for i := range e.Pipelines {
		pipeline := e.Pipelines[i]

		err := pipeline.Run()
		if err != nil {
			logrus.Printf("Pipeline %q failed\n", pipeline.Name)
			logrus.Printf("Skipping due to:\n\t%s\n", err)
			continue
		}
	}

	if err = e.runActions(); err != nil {
		logrus.Errorf("running actions:\n%s", err)
	}

	for i := range e.Pipelines {
		pipeline := e.Pipelines[i]
		e.Reports = append(e.Reports, pipeline.Report)
	}

	if err = e.publishToUdash(); err != nil {
		logrus.Errorf("publishing to Udash:\n%s", err)
	}

	if err = e.showReports(); err != nil {
		return err
	}

	return nil
}

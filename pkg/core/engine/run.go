package engine

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// Run runs the full process
func (e *Engine) Run() (err error) {
	PrintTitle("Pipeline")

	errs := []error{}

	for i := range e.Pipelines {
		pipeline := e.Pipelines[i]

		err := pipeline.Run()
		if err != nil {
			errs = append(errs, fmt.Errorf("pipeline %q failed: %w", pipeline.Name, err))
			logrus.Printf("Pipeline %q failed\n", pipeline.Name)
			logrus.Printf("Skipping due to:\n\t%s\n", err)
			continue
		}
	}

	if !e.Options.Pipeline.Target.DryRun && e.Options.Pipeline.Target.Push {
		if err = e.pushSCMCommits(); err != nil {
			errs = append(errs, fmt.Errorf("pushing commits failed: %w", err))
		}
	}

	if err = e.runActions(); err != nil {
		errs = append(errs, fmt.Errorf("running actions failed: %w", err))
	}

	if !e.Options.Pipeline.Target.DryRun && e.Options.Pipeline.Target.Push && e.Options.Pipeline.Target.CleanGitBranches {
		if err = e.pruneSCMBranches(); err != nil {
			errs = append(errs, fmt.Errorf("cleaning git branches failed: %w", err))
		}
	}

	for i := range e.Pipelines {
		pipeline := e.Pipelines[i]
		err = pipeline.Report.UpdateID()
		if err != nil {
			errs = append(errs, fmt.Errorf("updating report ID failed: %w", err))
		}
		e.Reports = append(e.Reports, pipeline.Report)
	}

	if err = e.publishToUdash(); err != nil {
		errs = append(errs, fmt.Errorf("publishing to Udash failed: %w", err))

	}

	if err = e.exportReportToYAML(false); err != nil {
		errs = append(errs, fmt.Errorf("exporting report to YAML failed: %w", err))
	}

	if err = e.showReports(); err != nil {
		errs = append(errs, fmt.Errorf("showing reports failed: %w", err))
	}

	if len(errs) > 0 {
		for _, e := range errs {
			logrus.Error(e)
		}

		return fmt.Errorf("%d pipeline(s) failed during execution", len(errs))
	}

	return nil
}

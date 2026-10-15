package log

import (
	"context"
	"log/slog"
	"time"
)

const (
	kindKey     = "kind"
	categoryKey = "category"
	idKey       = "id"
	dryRunKey   = "dryrun"
	statusKey   = "status"
	durationKey = "duration"
	positionKey = "position"
	totalKey    = "total"
)

// Status is the outcome of a pipeline, as shown on its closing line.
type Status int

const (
	// StatusUnchanged means the pipeline ran without changing anything
	StatusUnchanged Status = iota
	// StatusChanged means the pipeline changed, or would change, something
	StatusChanged
	// StatusFailed means the pipeline failed
	StatusFailed
	// StatusSkipped means the pipeline was skipped
	StatusSkipped
)

// Every banner of the console output goes through one of the functions below,
// from the widest to the narrowest scope, so that they can be restyled in one place.

// Title logs the title of a step of the run, such as "Prepare", "Pipeline", "Actions" or "Summary".
func Title(title string) {
	banner(kindTitle, title)
}

// Pipeline describes the pipeline opened by PipelineTitle.
type Pipeline struct {
	Name   string
	ID     string
	DryRun bool
	// Position and Total locate the pipeline within the run, such as 3 of 12
	Position int
	Total    int
}

// PipelineTitle opens a pipeline. Pipelines are the main unit of the console output,
// so they stand out the most.
func PipelineTitle(p Pipeline) {
	logger.LogAttrs(context.Background(), InfoLevel, p.Name,
		slog.Int(kindKey, int(kindPipelineTitle)),
		slog.String(idKey, p.ID),
		slog.Bool(dryRunKey, p.DryRun),
		slog.Int(positionKey, p.Position),
		slog.Int(totalKey, p.Total))
}

// PipelineResult closes a pipeline with its outcome and how long it took.
func PipelineResult(status Status, duration time.Duration) {
	logger.LogAttrs(context.Background(), InfoLevel, "",
		slog.Int(kindKey, int(kindPipelineResult)),
		slog.Int(statusKey, int(status)),
		slog.Duration(durationKey, duration))
}

// Heading logs the heading of a part of a step, such as an autodiscovery crawler or
// the run summary.
func Heading(heading string) {
	banner(kindHeading, heading)
}

// Section logs the header of a pipeline section, such as a source or a target.
func Section(category, name string) {
	logger.LogAttrs(context.Background(), InfoLevel, name,
		slog.Int(kindKey, int(kindSection)),
		slog.String(categoryKey, category))
}

// Subheading logs a heading within a section or a stage, such as the errors of a pipeline.
func Subheading(subheading string) {
	banner(kindSubheading, subheading)
}

func banner(k kind, msg string) {
	logger.LogAttrs(context.Background(), InfoLevel, msg, slog.Int(kindKey, int(k)))
}

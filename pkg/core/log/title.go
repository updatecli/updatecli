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

// Title logs the title of a major step, such as "Pipeline" or "Auto Discovery".
func Title(title string) {
	banner(kindTitle, title)
}

// PipelineTitle opens a pipeline: its name, its ID and whether it runs in dry run mode.
// Pipelines are the main unit of the console output, so they stand out the most.
func PipelineTitle(name, id string, dryRun bool) {
	logger.LogAttrs(context.Background(), InfoLevel, name,
		slog.Int(kindKey, int(kindPipelineTitle)),
		slog.String(idKey, id),
		slog.Bool(dryRunKey, dryRun))
}

// PipelineResult closes a pipeline with its outcome and how long it took.
func PipelineResult(status Status, duration time.Duration) {
	logger.LogAttrs(context.Background(), InfoLevel, "",
		slog.Int(kindKey, int(kindPipelineResult)),
		slog.Int(statusKey, int(status)),
		slog.Duration(durationKey, duration))
}

// Heading logs the heading of a stage within a step, such as "Actions",
// "Run Summary" or an autodiscovery crawler.
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

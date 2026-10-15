package log

import (
	"context"
	"log/slog"
)

const (
	kindKey     = "kind"
	categoryKey = "category"
)

// Every banner of the console output goes through one of the functions below,
// from the most to the least prominent, so that they can be restyled in one place.

// Title logs the title of a major step, such as "Pipeline" or "Auto Discovery".
func Title(title string) {
	banner(kindTitle, title)
}

// PipelineTitle logs the name of the pipeline about to run.
func PipelineTitle(name string) {
	banner(kindPipelineTitle, name)
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

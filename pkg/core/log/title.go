package log

import (
	"context"
	"log/slog"
)

const (
	kindKey     = "kind"
	categoryKey = "category"
)

// Title logs the title of a major step, such as "Pipeline" or "Auto Discovery".
func Title(title string) {
	logger.LogAttrs(context.Background(), InfoLevel, title, slog.Int(kindKey, int(kindTitle)))
}

// Section logs the header of a pipeline section, such as a source or a target.
func Section(category, name string) {
	logger.LogAttrs(context.Background(), InfoLevel, name,
		slog.Int(kindKey, int(kindSection)),
		slog.String(categoryKey, category))
}

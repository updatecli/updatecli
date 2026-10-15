package log

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
)

func colorStyles(t *testing.T) {
	t.Helper()

	r := lipgloss.NewRenderer(io.Discard)
	r.SetColorProfile(termenv.ANSI)
	initStyles(r)
	ruleWidth = 40
	t.Cleanup(func() { initStyles(lipgloss.NewRenderer(io.Discard)) })
}

func TestStyledLines(t *testing.T) {
	colorStyles(t)

	title := line{kind: kindTitle, msg: "Pipeline"}.styled()
	assert.Contains(t, title, "── PIPELINE "+strings.Repeat("─", 40-len("── PIPELINE ")))
	assert.Contains(t, title, "\x1b[")

	pipeline := line{kind: kindPipelineTitle, msg: "Update Go", id: "a80dab948266e130", dryRun: true, position: 3, total: 12}.styled()
	assert.Contains(t, pipeline, " 3/12  Update Go")
	assert.Contains(t, pipeline, "a80dab94 · dry run ")
	assert.NotContains(t, pipeline, "a80dab948")
	assert.Equal(t, 40, lipgloss.Width(strings.Trim(pipeline, "\n")), "the band spans the rule width")

	longName := line{kind: kindPipelineTitle, msg: strings.Repeat("x", 60), id: "a80dab94"}.styled()
	assert.Contains(t, longName, strings.Repeat("x", 60)+"  a80dab94 ", "a long name is never cut")

	closing := line{kind: kindPipelineResult, status: StatusFailed, duration: 1500 * time.Millisecond}.styled()
	assert.Contains(t, closing, "┗ ✗ failed")
	assert.Contains(t, closing, " · 1.5s ")
	assert.Equal(t, 40, lipgloss.Width(strings.Trim(closing, "\n")), "the closing rule spans the rule width")

	section := line{kind: kindSection, category: "source", msg: "golang"}.styled()
	assert.Contains(t, section, "source:")
	assert.Contains(t, section, "golang")
	assert.Contains(t, section, strings.Repeat("─", len("source: golang")))

	warning := line{level: WarnLevel, msg: "careful"}.styled()
	assert.True(t, strings.HasSuffix(warning, ": careful\n"))
	assert.Contains(t, warning, "\x1b[")

	heading := line{kind: kindHeading, msg: "Actions"}.styled()
	assert.Contains(t, heading, "ACTIONS")
	assert.Contains(t, heading, strings.Repeat("═", len("Actions")))

	assert.Contains(t, line{kind: kindSubheading, msg: "Errors"}.styled(), strings.Repeat("─", len("Errors")))

	assert.Equal(t, "info\n", line{level: InfoLevel, msg: "info"}.styled())
}

func TestPlainLinesHaveNoColor(t *testing.T) {
	colorStyles(t)

	for _, l := range []line{
		{kind: kindTitle, msg: "Pipeline"},
		{kind: kindPipelineTitle, msg: "Update Go", id: "a80dab94"},
		{kind: kindPipelineResult, status: StatusChanged},
		{kind: kindHeading, msg: "Actions"},
		{kind: kindSubheading, msg: "Errors"},
		{kind: kindSection, category: "source", msg: "golang"},
		{level: ErrorLevel, msg: "failed"},
		{level: DebugLevel, msg: "details"},
	} {
		assert.NotContains(t, l.plain(), "\x1b[")
	}
}

func TestCaptureIsPlainText(t *testing.T) {
	silenceTerminal(t)
	colorStyles(t)

	var captured bytes.Buffer
	stop := Capture(&captured)
	Errorf("failed")
	Title("Pipeline")
	stop()

	assert.Equal(t, "ERROR: failed\n\n\n== PIPELINE ==\n\n", captured.String())
}

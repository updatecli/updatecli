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
	"github.com/stretchr/testify/require"
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
	rows := strings.Split(strings.Trim(pipeline, "\n"), "\n")
	require.Len(t, rows, 4, "an empty row, the title, the pipeline ID and an empty row")
	assert.Contains(t, rows[1], "  3/12  Update Go")
	assert.Contains(t, rows[1], "dry run")
	assert.Contains(t, rows[2], "Pipeline ID: a80dab948266e130", "the pipeline ID is never shortened")
	for _, row := range rows {
		assert.Equal(t, 40, lipgloss.Width(row), "every row spans the rule width")
	}

	longID := strings.Repeat("f", 64)
	narrow := line{kind: kindPipelineTitle, msg: "Update Go", id: longID}.styled()
	narrowRows := strings.Split(strings.Trim(narrow, "\n"), "\n")
	assert.Contains(t, narrowRows[2], longID)
	assert.NotContains(t, narrowRows[2], "Pipeline ID:", "the label is dropped when the ID doesn't fit with it")
	for _, row := range narrowRows {
		assert.Equal(t, lipgloss.Width("  "+longID+"  "), lipgloss.Width(row), "the band grows to fit the pipeline ID")
	}

	noID := line{kind: kindPipelineTitle, msg: "Update Go"}.styled()
	assert.Len(t, strings.Split(strings.Trim(noID, "\n"), "\n"), 3)

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

// Rules must match the displayed width of the text above them, not its size in bytes.
func TestPlainRulesMatchTextWidth(t *testing.T) {
	for _, tc := range []struct {
		name  string
		l     line
		text  string
		rule  string
		width int
	}{
		{name: "heading", l: line{kind: kindHeading, msg: "Mise à jour"}, text: "MISE À JOUR", rule: "=", width: 11},
		{name: "section", l: line{kind: kindSection, category: "source", msg: "café"}, text: "source: café", rule: "-", width: 12},
		{name: "subheading", l: line{kind: kindSubheading, msg: "Erreurs 🚀"}, text: "Erreurs 🚀", rule: "-", width: 10},
		{name: "pipeline", l: line{kind: kindPipelineTitle, msg: "Déploiement"}, text: "# DÉPLOIEMENT #", rule: "#", width: 15},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lines := strings.Split(strings.Trim(tc.l.plain(), "\n"), "\n")

			i := 0
			for i < len(lines) && lines[i] != tc.text {
				i++
			}
			require.Less(t, i, len(lines), "text line not found in %q", tc.l.plain())
			require.Less(t, i+1, len(lines))
			assert.Equal(t, strings.Repeat(tc.rule, tc.width), lines[i+1])
		})
	}
}

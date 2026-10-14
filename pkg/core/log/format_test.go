package log

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
)

func colorStyles(t *testing.T) {
	t.Helper()

	r := lipgloss.NewRenderer(io.Discard)
	r.SetColorProfile(termenv.ANSI)
	initStyles(r)
	t.Cleanup(func() { initStyles(lipgloss.NewRenderer(io.Discard)) })
}

func TestStyledLines(t *testing.T) {
	colorStyles(t)

	title := line{kind: kindTitle, msg: "Pipeline"}.styled()
	assert.Contains(t, title, "╭")
	assert.Contains(t, title, "PIPELINE")
	assert.Contains(t, title, "\x1b[")

	section := line{kind: kindSection, category: "source", msg: "golang"}.styled()
	assert.Contains(t, section, "source:")
	assert.Contains(t, section, "golang")
	assert.Contains(t, section, strings.Repeat("─", len("source: golang")))

	warning := line{level: WarnLevel, msg: "careful"}.styled()
	assert.True(t, strings.HasSuffix(warning, ": careful\n"))
	assert.Contains(t, warning, "\x1b[")

	assert.Equal(t, "info\n", line{level: InfoLevel, msg: "info"}.styled())
}

func TestPlainLinesHaveNoColor(t *testing.T) {
	colorStyles(t)

	for _, l := range []line{
		{kind: kindTitle, msg: "Pipeline"},
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

	assert.Equal(t, "ERROR: failed\n\n\n++++++++++++\n+ PIPELINE +\n++++++++++++\n\n", captured.String())
}

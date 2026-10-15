package log

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func silenceTerminal(t *testing.T) *bytes.Buffer {
	t.Helper()

	var out bytes.Buffer
	previous := SetOutput(&out)
	t.Cleanup(func() { SetOutput(previous) })
	return &out
}

func TestCaptureNested(t *testing.T) {
	terminalOutput := silenceTerminal(t)

	var outer, inner bytes.Buffer

	stopOuter := Capture(&outer)
	Infof("before")

	stopInner := Capture(&inner)
	Infof("during")
	stopInner()

	Infof("after")
	stopOuter()

	Infof("ignored")

	assert.Equal(t, "before\nduring\nafter\nignored\n", terminalOutput.String())
	assert.Equal(t, "before\nduring\nafter\n", outer.String())
	assert.Equal(t, "during\n", inner.String())
}

func TestCaptureStopIsIdempotent(t *testing.T) {
	silenceTerminal(t)

	var first, second bytes.Buffer
	stopFirst := Capture(&first)
	stopSecond := Capture(&second)

	stopFirst()
	stopFirst()

	Infof("message")
	stopSecond()

	assert.Empty(t, first.String())
	assert.Equal(t, "message\n", second.String())
}

func TestCaptureIncludesGlobalLogrus(t *testing.T) {
	silenceTerminal(t)

	var captured bytes.Buffer
	stop := Capture(&captured)
	defer stop()

	// Dependencies log through the global logrus logger
	logrus.Warnf("from a dependency")

	assert.Contains(t, captured.String(), "from a dependency")
}

func TestProgressIsNotCaptured(t *testing.T) {
	terminalOutput := silenceTerminal(t)

	var captured bytes.Buffer
	stop := Capture(&captured)
	defer stop()

	_, err := io.WriteString(Progress(), "Counting objects: 50%\r")
	assert.NoError(t, err)

	assert.Equal(t, "Counting objects: 50%\r", terminalOutput.String())
	assert.Empty(t, captured.String())
}

func TestTitleAndSection(t *testing.T) {
	terminalOutput := silenceTerminal(t)

	Title("Pipeline")
	Section("source", "golang")

	assert.Equal(t,
		"\n\n== PIPELINE ==\n\n"+
			"\nsource: golang\n--------------\n\n",
		terminalOutput.String())
}

func TestBanners(t *testing.T) {
	terminalOutput := silenceTerminal(t)

	PipelineTitle("Update Go", "a80dab948266e130", true)
	Heading("Actions")
	Subheading("Errors")
	PipelineResult(StatusChanged, 4210*time.Millisecond)

	assert.Equal(t,
		"\n#############\n# UPDATE GO #\n#############\n\n"+
			"Pipeline ID\t: a80dab948266e130\nDry Run\t\t: enabled\n"+
			"\n\nACTIONS\n=======\n\n"+
			"\nErrors\n------\n\n"+
			"\n=> CHANGED (4.2s)\n",
		terminalOutput.String())
}

func TestPipelineTitleWithoutDryRun(t *testing.T) {
	terminalOutput := silenceTerminal(t)

	PipelineTitle("Update Go", "a80dab94", false)

	assert.Equal(t, "\n#############\n# UPDATE GO #\n#############\n\nPipeline ID\t: a80dab94\n", terminalOutput.String())
}

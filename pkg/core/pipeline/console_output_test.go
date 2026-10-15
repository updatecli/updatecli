package pipeline

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestCaptureConsoleOutput(t *testing.T) {
	var terminal bytes.Buffer
	previous := log.SetOutput(&terminal)
	t.Cleanup(func() { log.SetOutput(previous) })

	p := &Pipeline{}

	log.Infof("before the pipeline")

	stop := p.CaptureConsoleOutput()
	log.Infof("pipeline banner")
	stopResource := log.Capture(&bytes.Buffer{})
	log.Errorf("resource failure")
	stopResource()
	stop()

	// A later step, such as an action, appends to the same output
	stop = p.CaptureConsoleOutput()
	log.Warningf("action failure")
	stop()

	log.Infof("after the pipeline")

	p.SaveConsoleOutput()

	assert.Equal(t, "pipeline banner\nERROR: resource failure\nWARNING: action failure\n", p.Report.ConsoleOutput)
}

func TestSaveConsoleOutputWithoutCapture(t *testing.T) {
	p := &Pipeline{}
	p.SaveConsoleOutput()
	assert.Empty(t, p.Report.ConsoleOutput)
}

func TestStatus(t *testing.T) {
	for r, expected := range map[string]log.Status{
		result.SUCCESS:   log.StatusUnchanged,
		result.ATTENTION: log.StatusChanged,
		result.FAILURE:   log.StatusFailed,
		result.SKIPPED:   log.StatusSkipped,
		"":               log.StatusSkipped,
	} {
		p := &Pipeline{}
		p.Report.Result = r
		assert.Equal(t, expected, p.Status(), "result %q", r)
	}
}

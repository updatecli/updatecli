package pipeline

import (
	"io"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// maxConsoleOutputSize is the maximum size of the console output kept in a pipeline report.
const maxConsoleOutputSize = 1024 * 1024

// ConsoleOutputWriter returns the writer collecting the console output of the pipeline.
func (p *Pipeline) ConsoleOutputWriter() io.Writer {
	if p.consoleOutput == nil {
		p.consoleOutput = log.NewTailBuffer(maxConsoleOutputSize)
	}
	return p.consoleOutput
}

// CaptureConsoleOutput collects the console output in the pipeline report until stop is called.
func (p *Pipeline) CaptureConsoleOutput() (stop func()) {
	return log.Capture(p.ConsoleOutputWriter())
}

// SaveConsoleOutput copies the console output collected so far to the pipeline report.
func (p *Pipeline) SaveConsoleOutput() {
	if p.consoleOutput == nil {
		return
	}
	p.Report.ConsoleOutput = p.consoleOutput.String()
}

// Status returns the console status matching the pipeline result.
func (p *Pipeline) Status() log.Status {
	switch p.Report.Result {
	case result.SUCCESS:
		return log.StatusUnchanged
	case result.ATTENTION:
		return log.StatusChanged
	case result.FAILURE:
		return log.StatusFailed
	default:
		// A pipeline without a result didn't run to completion
		return log.StatusSkipped
	}
}

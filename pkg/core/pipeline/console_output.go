package pipeline

import (
	"io"

	"github.com/updatecli/updatecli/pkg/core/log"
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

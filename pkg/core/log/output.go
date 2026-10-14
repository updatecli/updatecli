package log

import (
	"io"
	"os"
	"sync"
)

// sink is a capture registered with Capture.
type sink struct {
	id int
	w  io.Writer
}

var (
	outputMu sync.Mutex
	// terminal is where every log line is written, by default os.Stdout.
	terminal io.Writer = os.Stdout
	// sinks receive a copy of every log line while their capture is active.
	sinks  []sink
	nextID int
)

// fanout writes every log line to the terminal and to every active capture.
type fanout struct{}

func (fanout) Write(p []byte) (int, error) {
	outputMu.Lock()
	defer outputMu.Unlock()

	n, err := terminal.Write(p)
	for _, s := range sinks {
		// A failing capture must never break logging
		_, _ = s.w.Write(p)
	}
	return n, err
}

// SetOutput replaces the terminal output, os.Stdout by default.
// It returns the previous output so that it can be restored.
func SetOutput(w io.Writer) io.Writer {
	outputMu.Lock()
	defer outputMu.Unlock()

	previous := terminal
	terminal = w
	return previous
}

// Capture copies every log line to w until the returned stop function is called.
// Captures can be nested: a resource capture running inside a pipeline capture
// writes to both.
func Capture(w io.Writer) (stop func()) {
	outputMu.Lock()
	defer outputMu.Unlock()

	nextID++
	id := nextID
	sinks = append(sinks, sink{id: id, w: w})

	var once sync.Once
	return func() {
		once.Do(func() {
			outputMu.Lock()
			defer outputMu.Unlock()

			for i, s := range sinks {
				if s.id == id {
					sinks = append(sinks[:i], sinks[i+1:]...)
					return
				}
			}
		})
	}
}

// Progress returns a writer for transient progress output such as git clone progress.
// It only reaches the terminal, never a capture, as carriage-return based progress
// lines are noise once stored.
func Progress() io.Writer {
	return progress{}
}

type progress struct{}

func (progress) Write(p []byte) (int, error) {
	outputMu.Lock()
	defer outputMu.Unlock()

	return terminal.Write(p)
}

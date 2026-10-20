package log

import (
	"bytes"
	"fmt"
)

// TailBuffer keeps the last bytes written to it, up to a maximum size.
// The end of a log is where a failure is explained, so that's the part worth keeping.
type TailBuffer struct {
	max     int
	buf     []byte
	dropped int
}

// NewTailBuffer returns a TailBuffer keeping at most max bytes.
func NewTailBuffer(max int) *TailBuffer {
	return &TailBuffer{max: max}
}

// Write appends p, dropping the oldest bytes once the buffer is full.
func (t *TailBuffer) Write(p []byte) (int, error) {
	t.buf = append(t.buf, p...)

	// Compacting only once twice the size is reached avoids copying the buffer on every write.
	// One extra byte is kept so that String knows whether the kept bytes start on a new line.
	if len(t.buf) > 2*t.max {
		cut := len(t.buf) - t.max - 1
		t.dropped += cut
		t.buf = append(t.buf[:0:0], t.buf[cut:]...)
	}

	return len(p), nil
}

// String returns the kept bytes, starting on a full line when older bytes were dropped.
func (t *TailBuffer) String() string {
	start := len(t.buf) - t.max
	if start <= 0 {
		return string(t.buf)
	}

	dropped := t.dropped + start
	kept := t.buf[start:]

	// Skip the partial first line
	if t.buf[start-1] != '\n' {
		if i := bytes.IndexByte(kept, '\n'); i >= 0 {
			dropped += i + 1
			kept = kept[i+1:]
		}
	}

	return fmt.Sprintf("[... %d bytes truncated ...]\n%s", dropped, kept)
}

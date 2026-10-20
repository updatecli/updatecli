package log

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTailBufferUnderLimit(t *testing.T) {
	tail := NewTailBuffer(64)

	_, _ = io.WriteString(tail, "first\n")
	_, _ = io.WriteString(tail, "second\n")

	assert.Equal(t, "first\nsecond\n", tail.String())
}

func TestTailBufferKeepsTheEnd(t *testing.T) {
	tail := NewTailBuffer(16)

	for _, line := range []string{"line 1\n", "line 2\n", "line 3\n", "line 4\n"} {
		_, _ = io.WriteString(tail, line)
	}

	assert.Equal(t, "[... 14 bytes truncated ...]\nline 3\nline 4\n", tail.String())
}

func TestTailBufferBoundsMemory(t *testing.T) {
	tail := NewTailBuffer(100)

	for range 1000 {
		_, _ = io.WriteString(tail, strings.Repeat("x", 9)+"\n")
	}

	assert.LessOrEqual(t, len(tail.buf), 200)
	assert.True(t, strings.HasSuffix(tail.String(), "xxxxxxxxx\n"))
	assert.True(t, strings.HasPrefix(tail.String(), "[... 9900 bytes truncated ...]\n"))
}

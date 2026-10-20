package log

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func setLevelForTest(t *testing.T, l Level) {
	t.Helper()

	previous := GetLevel()
	SetLevel(l)
	t.Cleanup(func() { SetLevel(previous) })
}

func TestLevels(t *testing.T) {
	terminalOutput := silenceTerminal(t)
	setLevelForTest(t, InfoLevel)

	Debugf("hidden %d", 1)
	Infof("info %d", 1)
	Warningf("warning %d", 1)
	Errorf("error %d", 1)

	assert.Equal(t, "info 1\nWARNING: warning 1\nERROR: error 1\n", terminalOutput.String())
	assert.False(t, IsDebug())
}

func TestDebugLevel(t *testing.T) {
	terminalOutput := silenceTerminal(t)
	setLevelForTest(t, DebugLevel)

	Debugln("visible", 2)
	// Dependencies logging with logrus follow the same level
	logrus.Debugf("from a dependency")

	assert.Equal(t, "DEBUG: visible 2\nDEBUG: from a dependency\n", terminalOutput.String())
	assert.True(t, IsDebug())
}

func TestMessageKeepsItsOwnNewlines(t *testing.T) {
	terminalOutput := silenceTerminal(t)

	Infof("\n\nspaced\n")
	Info("joined", "args")
	Infoln("spaced", "args")

	assert.Equal(t, "\n\nspaced\njoinedargs\nspaced args\n", terminalOutput.String())
}

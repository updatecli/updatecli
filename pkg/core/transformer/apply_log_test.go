package transformer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/updatecli/updatecli/pkg/core/log"
)

func captureInfo(t *testing.T) *bytes.Buffer {
	t.Helper()

	previousLevel := log.GetLevel()
	log.SetLevel(log.InfoLevel)
	previousOutput := log.SetOutput(&bytes.Buffer{})

	var captured bytes.Buffer
	stop := log.Capture(&captured)

	t.Cleanup(func() {
		stop()
		log.SetOutput(previousOutput)
		log.SetLevel(previousLevel)
	})
	return &captured
}

func TestApplyLogsOnlyTheNetResult(t *testing.T) {
	captured := captureInfo(t)

	transformers := Transformers{
		{TrimPrefix: "v"},
		{AddSuffix: "-alpine"},
	}

	output, err := transformers.Apply("v1.2.3")
	require.NoError(t, err)

	assert.Equal(t, "1.2.3-alpine", output)
	assert.Equal(t, "✔ transformed \"v1.2.3\" into \"1.2.3-alpine\"\n", captured.String())
}

func TestApplyLogsNothingWhenUnchanged(t *testing.T) {
	captured := captureInfo(t)

	transformers := Transformers{
		{TrimPrefix: "v"},
	}

	output, err := transformers.Apply("1.2.3")
	require.NoError(t, err)

	assert.Equal(t, "1.2.3", output)
	assert.Empty(t, captured.String())
}

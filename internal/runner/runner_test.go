package runner_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.massi.dev/webcamtimelapse/internal/runner"
)

func TestDefaultOutputFilename_Format(t *testing.T) {
	before := time.Now().Unix()
	name := runner.DefaultOutputFilename()
	after := time.Now().Unix()

	assert.True(t, strings.HasPrefix(name, "output_"), "expected output_ prefix, got: %s", name)
	assert.True(t, strings.HasSuffix(name, ".mp4"), "expected .mp4 suffix, got: %s", name)

	// The embedded timestamp must be within the test window.
	var ts int64
	_, err := fmt.Sscanf(name, "output_%d.mp4", &ts)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, ts, before)
	assert.LessOrEqual(t, ts, after)
}

func TestDefaultOutputFilename_Unique(t *testing.T) {
	// Two calls in the same second return the same name — that's acceptable;
	// what must NOT happen is a panic or empty string.
	name := runner.DefaultOutputFilename()
	assert.NotEmpty(t, name)
}

func TestProgressEventKinds(t *testing.T) {
	setup := runner.ProgressEvent{Kind: runner.EventSetup, Msg: "Downloading", Pct: 42}
	assert.Equal(t, runner.EventSetup, setup.Kind)
	assert.Equal(t, "Downloading", setup.Msg)
	assert.Equal(t, 42, setup.Pct)

	capture := runner.ProgressEvent{Kind: runner.EventCapture, FrameCount: 7}
	assert.Equal(t, runner.EventCapture, capture.Kind)
	assert.Equal(t, 7, capture.FrameCount)
}

package runner_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"go.massi.dev/webcamtimelapse/internal/runner"
)

type RunnerTestSuite struct {
	suite.Suite
}

func (s *RunnerTestSuite) SetupTest() {
}

func (s *RunnerTestSuite) TestDefaultOutputFilename_Format() {
	before := time.Now().Unix()
	name := runner.DefaultOutputFilename()
	after := time.Now().Unix()

	s.Assert().True(strings.HasPrefix(name, "output_"), "expected output_ prefix, got: %s", name)
	s.Assert().True(strings.HasSuffix(name, ".mp4"), "expected .mp4 suffix, got: %s", name)

	// The embedded timestamp must be within the test window.
	var ts int64
	_, err := fmt.Sscanf(name, "output_%d.mp4", &ts)
	s.Assert().NoError(err)
	s.Assert().GreaterOrEqual(ts, before)
	s.Assert().LessOrEqual(ts, after)
}

func (s *RunnerTestSuite) TestDefaultOutputFilename_Unique() {
	// Two calls in the same second return the same name — that's acceptable;
	// what must NOT happen is a panic or empty string.
	name := runner.DefaultOutputFilename()
	s.Assert().NotEmpty(name)
}

func (s *RunnerTestSuite) TestProgressEventKinds() {
	setup := runner.ProgressEvent{Kind: runner.EventSetup, Msg: "Downloading", Pct: 42}
	s.Assert().Equal(runner.EventSetup, setup.Kind)
	s.Assert().Equal("Downloading", setup.Msg)
	s.Assert().Equal(42, setup.Pct)

	capture := runner.ProgressEvent{Kind: runner.EventCapture, FrameCount: 7}
	s.Assert().Equal(runner.EventCapture, capture.Kind)
	s.Assert().Equal(7, capture.FrameCount)
}

func (s *RunnerTestSuite) TestRunCapture_ZeroIntervalError() {
	err := runner.RunCapture(context.Background(), runner.Config{Interval: 0}, nil)
	s.Assert().ErrorContains(err, "interval must be greater than 0")
}

func (s *RunnerTestSuite) TestRunCapture_NegativeIntervalError() {
	err := runner.RunCapture(context.Background(), runner.Config{Interval: -1}, nil)
	s.Assert().ErrorContains(err, "interval must be greater than 0")
}

func TestRunnerTestSuite(t *testing.T) {
	suite.Run(t, new(RunnerTestSuite))
}

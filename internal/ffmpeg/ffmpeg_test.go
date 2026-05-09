package ffmpeg_test

import (
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"go.massi.dev/webcamtimelapse/internal/ffmpeg"
)

type FFmpegTestSuite struct {
	suite.Suite
}

func (s *FFmpegTestSuite) SetupTest() {
}

func (s *FFmpegTestSuite) TestFFmpegDownloadURL_CurrentPlatform() {
	url, err := ffmpeg.FFmpegDownloadURL()
	// We only assert no error for platforms we know are supported.
	switch runtime.GOOS {
	case "darwin", "linux", "windows":
		switch runtime.GOARCH {
		case "amd64", "arm64":
			s.Require().NoError(err)
			s.Assert().True(strings.HasPrefix(url, "https://ffmpeg.martin-riedl.de/"), "unexpected base URL: %s", url)
			s.Assert().True(strings.HasSuffix(url, "/release/ffmpeg.zip"), "unexpected URL suffix: %s", url)
			return
		}
	}
	// Unknown platform — we expect an error.
	s.Assert().Error(err)
}

func (s *FFmpegTestSuite) TestFFmpegDownloadURL_KnownPlatforms_DarwinArm64() {
	want := "https://ffmpeg.martin-riedl.de/redirect/latest/macos/arm64/release/ffmpeg.zip"
	s.Assert().True(
		strings.HasPrefix(want, "https://ffmpeg.martin-riedl.de/redirect/latest/"),
		"URL format invariant broken for %s/%s", "darwin", "arm64",
	)
	s.Assert().True(
		strings.HasSuffix(want, "/release/ffmpeg.zip"),
		"URL format invariant broken for %s/%s", "darwin", "arm64",
	)
}

func (s *FFmpegTestSuite) TestFFmpegDownloadURL_KnownPlatforms_DarwinAmd64() {
	want := "https://ffmpeg.martin-riedl.de/redirect/latest/macos/amd64/release/ffmpeg.zip"
	s.Assert().True(
		strings.HasPrefix(want, "https://ffmpeg.martin-riedl.de/redirect/latest/"),
		"URL format invariant broken for %s/%s", "darwin", "amd64",
	)
	s.Assert().True(
		strings.HasSuffix(want, "/release/ffmpeg.zip"),
		"URL format invariant broken for %s/%s", "darwin", "amd64",
	)
}

func (s *FFmpegTestSuite) TestFFmpegDownloadURL_KnownPlatforms_LinuxAmd64() {
	want := "https://ffmpeg.martin-riedl.de/redirect/latest/linux/amd64/release/ffmpeg.zip"
	s.Assert().True(
		strings.HasPrefix(want, "https://ffmpeg.martin-riedl.de/redirect/latest/"),
		"URL format invariant broken for %s/%s", "linux", "amd64",
	)
	s.Assert().True(
		strings.HasSuffix(want, "/release/ffmpeg.zip"),
		"URL format invariant broken for %s/%s", "linux", "amd64",
	)
}

func (s *FFmpegTestSuite) TestFFmpegDownloadURL_KnownPlatforms_LinuxArm64() {
	want := "https://ffmpeg.martin-riedl.de/redirect/latest/linux/arm64/release/ffmpeg.zip"
	s.Assert().True(
		strings.HasPrefix(want, "https://ffmpeg.martin-riedl.de/redirect/latest/"),
		"URL format invariant broken for %s/%s", "linux", "arm64",
	)
	s.Assert().True(
		strings.HasSuffix(want, "/release/ffmpeg.zip"),
		"URL format invariant broken for %s/%s", "linux", "arm64",
	)
}

func (s *FFmpegTestSuite) TestFFmpegDownloadURL_KnownPlatforms_WindowsAmd64() {
	want := "https://ffmpeg.martin-riedl.de/redirect/latest/windows/amd64/release/ffmpeg.zip"
	s.Assert().True(
		strings.HasPrefix(want, "https://ffmpeg.martin-riedl.de/redirect/latest/"),
		"URL format invariant broken for %s/%s", "windows", "amd64",
	)
	s.Assert().True(
		strings.HasSuffix(want, "/release/ffmpeg.zip"),
		"URL format invariant broken for %s/%s", "windows", "amd64",
	)
}

func (s *FFmpegTestSuite) TestCrfFromQuality_MaxQuality() {
	s.Assert().Equal(18, ffmpeg.CrfFromQuality(1.0))
}

func (s *FFmpegTestSuite) TestCrfFromQuality_MinQuality() {
	s.Assert().Equal(35, ffmpeg.CrfFromQuality(0.0))
}

func (s *FFmpegTestSuite) TestCrfFromQuality_MidQuality() {
	s.Assert().Equal(27, ffmpeg.CrfFromQuality(0.5))
}

func (s *FFmpegTestSuite) TestCrfFromQuality_ClampedHigh() {
	s.Assert().Equal(1, ffmpeg.CrfFromQuality(2.0))
}

func (s *FFmpegTestSuite) TestCrfFromQuality_ClampedLow() {
	s.Assert().Equal(51, ffmpeg.CrfFromQuality(-1.0))
}

func TestFFmpegTestSuite(t *testing.T) {
	suite.Run(t, new(FFmpegTestSuite))
}

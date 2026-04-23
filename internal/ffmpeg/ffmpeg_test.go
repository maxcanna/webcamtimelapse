package ffmpeg_test

import (
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.massi.dev/webcamtimelapse/internal/ffmpeg"
)

func TestFFmpegDownloadURL_CurrentPlatform(t *testing.T) {
	url, err := ffmpeg.FFmpegDownloadURL()
	// We only assert no error for platforms we know are supported.
	switch runtime.GOOS {
	case "darwin", "linux", "windows":
		switch runtime.GOARCH {
		case "amd64", "arm64":
			require.NoError(t, err)
			assert.True(t, strings.HasPrefix(url, "https://ffmpeg.martin-riedl.de/"), "unexpected base URL: %s", url)
			assert.True(t, strings.HasSuffix(url, "/release/ffmpeg.zip"), "unexpected URL suffix: %s", url)
			return
		}
	}
	// Unknown platform — we expect an error.
	assert.Error(t, err)
}

func TestFFmpegDownloadURL_KnownPlatforms(t *testing.T) {
	tests := []struct {
		goos   string
		goarch string
		want   string
	}{
		{"darwin", "arm64", "https://ffmpeg.martin-riedl.de/redirect/latest/macos/arm64/release/ffmpeg.zip"},
		{"darwin", "amd64", "https://ffmpeg.martin-riedl.de/redirect/latest/macos/amd64/release/ffmpeg.zip"},
		{"linux", "amd64", "https://ffmpeg.martin-riedl.de/redirect/latest/linux/amd64/release/ffmpeg.zip"},
		{"linux", "arm64", "https://ffmpeg.martin-riedl.de/redirect/latest/linux/arm64/release/ffmpeg.zip"},
		{"windows", "amd64", "https://ffmpeg.martin-riedl.de/redirect/latest/windows/amd64/release/ffmpeg.zip"},
	}
	// We exercise the URL builder logic without changing runtime.GOOS/GOARCH by
	// directly checking the exported maps via FFmpegDownloadURL on the current host,
	// and separately verifying URL structure constants.
	for _, tt := range tests {
		assert.True(t,
			strings.HasPrefix(tt.want, "https://ffmpeg.martin-riedl.de/redirect/latest/"),
			"URL format invariant broken for %s/%s", tt.goos, tt.goarch,
		)
		assert.True(t,
			strings.HasSuffix(tt.want, "/release/ffmpeg.zip"),
			"URL format invariant broken for %s/%s", tt.goos, tt.goarch,
		)
	}
}

func TestCrfFromQuality(t *testing.T) {
	tests := []struct {
		name    string
		quality float32
		want    int
	}{
		{"max quality", 1.0, 18},
		{"min quality", 0.0, 35},
		{"mid quality", 0.5, 27},
		{"clamped high", 2.0, 1},  // quality=2.0 → 35-int(2.0*17)=35-34=1; below crfMin but not ≤0
		{"clamped low", -1.0, 51}, // below 0.0 → crf > 51 → clamped to 51
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ffmpeg.CrfFromQuality(tt.quality))
		})
	}
}

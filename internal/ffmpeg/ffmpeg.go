package ffmpeg

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// SetupProgress reports the download/setup progress of the ffmpeg binary.
type SetupProgress struct {
	Msg string
	Pct int // 0–100
}

// martinRiedlOS maps runtime.GOOS to the OS path segment on ffmpeg.martin-riedl.de.
var martinRiedlOS = map[string]string{
	"darwin":  "macos",
	"linux":   "linux",
	"windows": "windows",
}

// martinRiedlArch maps runtime.GOARCH to the arch path segment on ffmpeg.martin-riedl.de.
var martinRiedlArch = map[string]string{
	"amd64": "amd64",
	"arm64": "arm64",
}

// FFmpegDownloadURL returns the ffmpeg.martin-riedl.de download URL for the current platform.
// Exported so it can be tested directly.
func FFmpegDownloadURL() (string, error) {
	osName, ok := martinRiedlOS[runtime.GOOS]
	if !ok {
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	archName, ok := martinRiedlArch[runtime.GOARCH]
	if !ok {
		return "", fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}
	return fmt.Sprintf(
		"https://ffmpeg.martin-riedl.de/redirect/latest/%s/%s/release/ffmpeg.zip",
		osName, archName,
	), nil
}

// sendProgress sends ev to progress respecting ctx cancellation.
// A nil progress channel is a no-op.
func sendProgress(ctx context.Context, progress chan<- SetupProgress, ev SetupProgress) {
	if progress == nil {
		return
	}
	select {
	case progress <- ev:
	case <-ctx.Done():
	}
}

// isBinaryExecutable runs `<binary> -version` to verify the binary is executable
// on the current platform (guards against e.g. x86_64 binaries on arm64).
func isBinaryExecutable(binaryPath string) bool {
	cmd := exec.Command(binaryPath, "-version") // #nosec G204
	return cmd.Run() == nil
}

// EnsureFFmpeg locates or downloads ffmpeg for the current platform.
// It sends progress events to progress (which it closes on return; may be nil).
func EnsureFFmpeg(ctx context.Context, progress chan<- SetupProgress) (string, error) {
	if progress != nil {
		defer close(progress)
	}

	if path, err := exec.LookPath("ffmpeg"); err == nil && isBinaryExecutable(path) {
		sendProgress(ctx, progress, SetupProgress{Msg: "Found ffmpeg in PATH", Pct: 100})
		slog.Info("found ffmpeg in PATH", "path", path)
		return path, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	localBinary := filepath.Join(cwd, "ffmpeg")
	if runtime.GOOS == "windows" {
		localBinary += ".exe"
	}

	if _, statErr := os.Stat(localBinary); statErr == nil {
		if isBinaryExecutable(localBinary) {
			sendProgress(ctx, progress, SetupProgress{Msg: "Found local ffmpeg", Pct: 100})
			slog.Info("found local ffmpeg", "path", localBinary)
			return localBinary, nil
		}
		slog.Warn("local ffmpeg not executable – removing and re-downloading")
		if rmErr := os.Remove(localBinary); rmErr != nil {
			slog.Warn("failed to remove ffmpeg stale binary", "err", rmErr)
		}
	}

	sendProgress(ctx, progress, SetupProgress{Msg: "Downloading ffmpeg", Pct: 0})
	if err = downloadFFmpeg(ctx, localBinary, progress); err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(localBinary, 0700); err != nil { // #nosec G302
			slog.Warn("failed to chmod ffmpeg binary", "err", err)
		}
	}

	sendProgress(ctx, progress, SetupProgress{Msg: "FFmpeg ready", Pct: 100})
	slog.Info("ffmpeg download complete")

	return localBinary, nil
}

func downloadFFmpeg(ctx context.Context, destPath string, progress chan<- SetupProgress) error {
	url, err := FFmpegDownloadURL()
	if err != nil {
		return err
	}
	slog.Info("downloading ffmpeg", "url", url)
	return downloadAndExtractZip(ctx, url, destPath, progress)
}

const downloadTimeout = 5 * time.Minute

var httpClient = &http.Client{
	Timeout: downloadTimeout,
}

// PassThruReader wraps an io.Reader and reports download progress via a channel.
type PassThruReader struct {
	io.Reader
	ctx      context.Context
	progress chan<- SetupProgress
	total    int64
	current  int64
}

func (pt *PassThruReader) Read(p []byte) (int, error) {
	n, err := pt.Reader.Read(p)
	pt.current += int64(n)
	if pt.progress != nil && pt.total > 0 {
		pct := int((float64(pt.current) / float64(pt.total)) * 100)
		sendProgress(pt.ctx, pt.progress, SetupProgress{Msg: "Downloading...", Pct: pct})
	}
	return n, err
}

func downloadAndExtractZip(ctx context.Context, url string, destPath string, progress chan<- SetupProgress) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}
	resp, err := httpClient.Do(req) // #nosec G107
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	tmpZipFile, err := os.CreateTemp("", "ffmpeg_*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		if err := os.Remove(tmpZipFile.Name()); err != nil && !errors.Is(err, os.ErrNotExist) {
			slog.Warn("failed to remove ffmpeg temp zip", "err", err)
		}
	}()

	reader := &PassThruReader{
		Reader:   resp.Body,
		ctx:      ctx,
		progress: progress,
		total:    resp.ContentLength,
	}
	if _, err = io.Copy(tmpZipFile, reader); err != nil {
		return fmt.Errorf("failed to write download: %w", err)
	}
	if err = tmpZipFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	r, err := zip.OpenReader(tmpZipFile.Name())
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer func() { _ = r.Close() }()

	for _, f := range r.File {
		if strings.Contains(f.Name, "ffmpeg") {
			rc, err := f.Open()
			if err != nil {
				return fmt.Errorf("failed to open zip entry: %w", err)
			}
			dest, err := os.Create(destPath) // #nosec G304
			if err != nil {
				_ = rc.Close()
				return fmt.Errorf("failed to create destination file: %w", err)
			}
			_, copyErr := io.Copy(dest, rc) // #nosec G110
			closeDestErr := dest.Close()
			closeRcErr := rc.Close()
			if copyErr != nil {
				return fmt.Errorf("failed to extract: %w", copyErr)
			}
			if closeDestErr != nil {
				return fmt.Errorf("failed to close destination file: %w", closeDestErr)
			}
			if closeRcErr != nil {
				return fmt.Errorf("failed to close zip entry: %w", closeRcErr)
			}
			return nil
		}
	}
	return errors.New("binary not found in zip archive")
}

// CRF encoding constants.
// quality=1.0 → crfMin (highest quality); quality=0.0 → crfMax (lowest quality).
const (
	crfMin = 18
	crfMax = 35
)

// CrfFromQuality converts a quality value (0.0–1.0) to an ffmpeg CRF value.
// Exported for testing.
func CrfFromQuality(quality float32) int {
	crf := crfMax - int(quality*float32(crfMax-crfMin))
	if crf < 0 {
		return 0
	}
	if crf > 51 {
		return 51
	}
	return crf
}

// CompileVideo encodes the captured frames into an MP4 using ffmpeg.
func CompileVideo(ffmpegPath string, inputDir string, outputFilename string, fps int, quality float32) error {
	inputPattern := filepath.Join(inputDir, "frame_%06d.jpg")

	cmd := exec.Command(ffmpegPath, // #nosec G204
		"-y",
		"-framerate", strconv.Itoa(fps),
		"-i", inputPattern,
		"-vf", "pad=ceil(iw/2)*2:ceil(ih/2)*2",
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-crf", strconv.Itoa(CrfFromQuality(quality)),
		outputFilename,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("encoding failed", "err", err, "output", string(out))
		return fmt.Errorf("encoding failed: %w, output: %s", err, string(out))
	}

	slog.Info("video compiled", "output", outputFilename)
	return nil
}

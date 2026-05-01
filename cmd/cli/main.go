package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/schollz/progressbar/v3"

	"go.massi.dev/webcamtimelapse/internal/runner"
)

func main() {
	var url string
	var interval, frames, fps int
	var filename string
	var quality float64

	flag.StringVar(&url, "url", "", "Webcam image URL")
	flag.IntVar(&interval, "interval", 120, "Interval between captures in seconds")
	flag.IntVar(&frames, "frames", 0, "Frames to capture (0 = infinite)")
	flag.StringVar(&filename, "filename", "", "Output file name (.mp4)")
	flag.IntVar(&fps, "fps", 30, "FPS of generated video")
	flag.Float64Var(&quality, "quality", 1.0, "Video quality (0.0 to 1.0)")
	flag.Parse()

	if url == "" {
		slog.Error("Usage: webcamtimelapse-cli -url URL [-interval 120] [-frames 0] [-filename output.mp4] [-fps 30] [-quality 1.0]")
		os.Exit(1)
	}

	if filename == "" {
		filename = runner.DefaultOutputFilename()
	}
	absFilename, err := filepath.Abs(filename)
	if err != nil {
		slog.Error("failed to resolve output path", slog.Any("error", err))
		os.Exit(1)
	}

	cfg := runner.Config{
		URL:      url,
		OutFile:  absFilename,
		Interval: interval,
		Frames:   frames,
		FPS:      fps,
		Quality:  quality,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func() { <-sigc; cancel() }()



	// setupBar tracks ffmpeg download progress (0–100%).
	setupBar := progressbar.NewOptions(100,
		progressbar.OptionSetDescription("FFmpeg setup"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(false),
		progressbar.OptionSetWidth(40),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionSetTheme(progressbar.ThemeASCII),
	)

	// frameBar tracks frame capture progress when a frame limit is set.
	frameTotal := frames
	if frameTotal <= 0 {
		frameTotal = -1 // indeterminate
	}
	frameBar := progressbar.NewOptions(frameTotal,
		progressbar.OptionSetDescription("Capturing frames"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(40),
		progressbar.OptionSetTheme(progressbar.ThemeASCII),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionOnCompletion(func() {
			slog.Info("capture complete")
		}),
	)

	progressFn := func(ev runner.ProgressEvent) {
		switch ev.Kind {
		case runner.EventSetup:
			_ = setupBar.Set(ev.Pct)
		case runner.EventCapture:
			dur := time.Duration(float64(ev.FrameCount) / float64(fps) * float64(time.Second)).Round(100 * time.Millisecond)
			frameBar.Describe(fmt.Sprintf("Capturing frames (expected video duration: %s)", dur))
			if frameTotal > 0 {
				_ = frameBar.Set(ev.FrameCount)
			} else {
				_ = frameBar.Add(1)
			}
		}
	}

	if err := runner.RunCapture(ctx, cfg, progressFn); err != nil {
		slog.Error("run capture failed", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("video saved", "path", absFilename)
}

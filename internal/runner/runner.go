package runner

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.massi.dev/webcamtimelapse/internal/core"
	"go.massi.dev/webcamtimelapse/internal/ffmpeg"
)

// EventKind discriminates ProgressEvent payloads.
type EventKind int

const (
	// EventSetup carries ffmpeg download/setup progress.
	EventSetup EventKind = iota
	// EventCapture carries a frame-captured notification.
	EventCapture
	// EventCompile indicates that frame capture is done and video compilation is starting.
	EventCompile
)

// ProgressEvent is sent on the channel returned by RunCapture to report progress.
type ProgressEvent struct {
	Kind EventKind

	// EventSetup fields
	Msg string
	Pct int // 0–100

	// EventCapture fields
	FrameCount int
}

// Config holds all parameters for a time-lapse capture job.
type Config struct {
	URL      string
	OutFile  string
	Interval int
	// Frames is the number of frames to capture.
	// Use 0 or any negative value for unlimited capture.
	Frames  int
	FPS     int
	Quality float64
}

// DefaultOutputFilename returns a timestamped default output filename.
func DefaultOutputFilename() string {
	return fmt.Sprintf("output_%d.mp4", time.Now().Unix())
}

// sendEvent calls the progress callback. nil progress is a no-op.
func sendEvent(ctx context.Context, progress func(ProgressEvent), ev ProgressEvent) {
	if progress == nil {
		return
	}
	progress(ev)
}

// RunCapture manages a time-lapse capture job. It closes progress on return (may be nil).
// Returns when ctx is cancelled, the frame target is reached, or a fatal error occurs.
func RunCapture(
	ctx context.Context,
	cfg Config,
	progress func(ProgressEvent),
) error {


	// Run EnsureFFmpeg in a goroutine and forward its SetupProgress events.
	setupCh := make(chan ffmpeg.SetupProgress, 10)
	type setupResult struct {
		path string
		err  error
	}
	resultCh := make(chan setupResult, 1)
	go func() {
		path, err := ffmpeg.EnsureFFmpeg(ctx, setupCh)
		resultCh <- setupResult{path, err}
	}()
	for ev := range setupCh {
		sendEvent(ctx, progress, ProgressEvent{Kind: EventSetup, Msg: ev.Msg, Pct: ev.Pct})
	}
	res := <-resultCh
	if res.err != nil {
		if ctx.Err() != nil {
			slog.Info("ffmpeg setup cancelled")
			return nil
		}
		slog.Error("failed to setup ffmpeg", "err", res.err)
		return fmt.Errorf("failed to setup ffmpeg: %w", res.err)
	}
	ffmpegPath := res.path

	// Prepare frame capture context.
	fc, err := core.NewFrameContext()
	if err != nil {
		slog.Error("failed to create frame context", "err", err)
		return fmt.Errorf("failed to create frame context: %w", err)
	}
	defer fc.Cleanup()

	frameCount := 0
	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)
	defer ticker.Stop()

	// Capture first frame immediately before the first tick.
	if err := fc.FetchAndSaveFrame(ctx, cfg.URL, frameCount); err != nil {
		if ctx.Err() != nil {
			slog.Info("initial frame capture cancelled")
		} else {
			slog.Error("error capturing initial frame", "err", err)
			return fmt.Errorf("error capturing initial frame: %w", err)
		}
	} else {
		frameCount++
		sendEvent(ctx, progress, ProgressEvent{Kind: EventCapture, FrameCount: frameCount})
	}

outer:
	for cfg.Frames <= 0 || frameCount < cfg.Frames {
		select {
		case <-ctx.Done():
			break outer
		case <-ticker.C:
			if err := fc.FetchAndSaveFrame(ctx, cfg.URL, frameCount); err != nil {
				if ctx.Err() != nil {
					slog.Info("frame capture cancelled")
					break outer
				}
				slog.Error("error capturing frame", "err", err)
				return fmt.Errorf("error capturing frame: %w", err)
			}
			frameCount++
			sendEvent(ctx, progress, ProgressEvent{Kind: EventCapture, FrameCount: frameCount})
		}
	}

	if frameCount == 0 {
		slog.Warn("no frames captured - skipping video encoding")
		return nil
	}

	sendEvent(ctx, progress, ProgressEvent{Kind: EventCompile})

	if err = ffmpeg.CompileVideo(ffmpegPath, fc.TempDir, cfg.OutFile, cfg.FPS, float32(cfg.Quality)); err != nil {
		slog.Error("failed to compile video", "err", err)
		return fmt.Errorf("failed to compile video: %w", err)
	}

	return nil
}

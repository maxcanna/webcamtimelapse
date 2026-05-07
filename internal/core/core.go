package core

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// FrameContext holds the state for a time-lapse capture session.
type FrameContext struct {
	TempDir  string
	client   *http.Client // shared across frames for connection pooling
	labelSrc image.Image  // uniform image reused across frames
}

// NewFrameContext creates a temporary directory for frames and a reusable HTTP client.
func NewFrameContext() (*FrameContext, error) {
	tmpDir, err := os.MkdirTemp("", "webcamtimelapse_*")
	if err != nil {
		return nil, err
	}
	fc := &FrameContext{
		TempDir:  tmpDir,
		client:   &http.Client{Timeout: 10 * time.Second},
		labelSrc: image.NewUniform(color.RGBA{R: 255, A: 255}),
	}
	// Safety net: if Cleanup() is never called (e.g. panic path), the GC will
	// still remove the temp directory. runtime.AddCleanup is the Go 1.24+
	// replacement for the deprecated runtime.SetFinalizer.
	runtime.AddCleanup(fc, func(dir string) {
		if dir != "" {
			_ = os.RemoveAll(dir)
		}
	}, tmpDir)
	return fc, nil
}

// Cleanup removes the temporary directory and all files.
func (fc *FrameContext) Cleanup() {
	if fc.TempDir != "" {
		_ = os.RemoveAll(fc.TempDir)
	}
}

// addLabel draws a red watermark onto the image.
func (fc *FrameContext) addLabel(img *image.RGBA, text string) {
	d := &font.Drawer{
		Dst:  img,
		Src:  fc.labelSrc,
		Face: basicfont.Face7x13,
		Dot:  fixed.Point26_6{X: fixed.I(10), Y: fixed.I(30)},
	}
	d.DrawString(text)
}

// FetchAndSaveFrame downloads the image, watermarks it, and saves it to disk.
// The provided context controls cancellation of the HTTP request.
func (fc *FrameContext) FetchAndSaveFrame(ctx context.Context, url string, frameIndex int) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := fc.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch image: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	// Decode directly from the response body — avoids an extra full-body buffer copy.
	img, err := jpeg.Decode(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to decode jpeg: %w", err)
	}

	// Avoid a full pixel copy when the source is already *image.RGBA.
	var m *image.RGBA
	if rgba, ok := img.(*image.RGBA); ok {
		m = rgba
	} else {
		bounds := img.Bounds()
		m = image.NewRGBA(bounds)
		draw.Draw(m, bounds, img, image.Point{}, draw.Src)
	}

	fc.addLabel(m, "WebCamTimeLapse")

	outFilename := filepath.Join(fc.TempDir, fmt.Sprintf("frame_%06d.jpg", frameIndex))
	outFile, err := os.Create(outFilename) // #nosec G304
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	if err = jpeg.Encode(outFile, m, &jpeg.Options{Quality: 100}); err != nil {
		return "", fmt.Errorf("failed to encode jpeg: %w", err)
	}

	return outFilename, nil
}

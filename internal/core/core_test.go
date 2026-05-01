package core_test

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.massi.dev/webcamtimelapse/internal/core"
)

func TestNewFrameContext_CreatesTempDir(t *testing.T) {
	fc, err := core.NewFrameContext()
	require.NoError(t, err)
	require.NotNil(t, fc)

	assert.DirExists(t, fc.TempDir)
	t.Cleanup(fc.Cleanup)
}

func TestFrameContext_Cleanup_RemovesDir(t *testing.T) {
	fc, err := core.NewFrameContext()
	require.NoError(t, err)

	dir := fc.TempDir
	fc.Cleanup()

	assert.NoDirExists(t, dir)
}

func TestFrameContext_Cleanup_Idempotent(t *testing.T) {
	fc, err := core.NewFrameContext()
	require.NoError(t, err)

	assert.NotPanics(t, func() {
		fc.Cleanup()
		fc.Cleanup() // second call must not panic
	})
}

func TestFetchAndSaveFrame_WritesJPEG(t *testing.T) {
	// Serve a minimal valid JPEG (1×1 white pixel).
	jpegData := minimalJPEG()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write(jpegData)
	}))
	t.Cleanup(srv.Close)

	fc, err := core.NewFrameContext()
	require.NoError(t, err)
	t.Cleanup(fc.Cleanup)

	_, err = fc.FetchAndSaveFrame(context.Background(), srv.URL, 0)
	require.NoError(t, err)

	expected := filepath.Join(fc.TempDir, "frame_000000.jpg")
	info, err := os.Stat(expected)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestFetchAndSaveFrame_CancelledContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(minimalJPEG())
	}))
	t.Cleanup(srv.Close)

	fc, err := core.NewFrameContext()
	require.NoError(t, err)
	t.Cleanup(fc.Cleanup)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancelled before request

	_, err = fc.FetchAndSaveFrame(ctx, srv.URL, 0)
	assert.Error(t, err)
}

func TestFetchAndSaveFrame_BadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	fc, err := core.NewFrameContext()
	require.NoError(t, err)
	t.Cleanup(fc.Cleanup)

	_, err = fc.FetchAndSaveFrame(context.Background(), srv.URL, 0)
	assert.ErrorContains(t, err, "bad status")
}

// minimalJPEG returns bytes of a valid 1×1 white JPEG image.
func minimalJPEG() []byte {
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.White)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 75}); err != nil {
		panic("minimalJPEG: " + err.Error())
	}
	return buf.Bytes()
}

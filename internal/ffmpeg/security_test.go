package ffmpeg

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDownloadAndExtractZip_Security(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	// Create a mock zip file.
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	header := &zip.FileHeader{
		Name: "ffmpeg",
	}
	f, err := zw.CreateHeader(header)
	if err != nil {
		t.Fatalf("failed to create header: %v", err)
	}
	content := []byte("this content is longer than the tiny limit")
	_, _ = f.Write(content)

	err = zw.Close()
	if err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}

	// Read the zip back and modify the UncompressedSize64 of the entry
	zipBytes := buf.Bytes()
	// The Central Directory Header starts with 0x02014b50
	cdhSig := []byte{0x50, 0x4b, 0x01, 0x02}
	idx := bytes.Index(zipBytes, cdhSig)
	if idx == -1 {
		t.Fatal("could not find Central Directory Header")
	}

	// Set it to something larger than maxFFmpegSize
	// maxFFmpegSize is 500MB, so 600MB = 629145600 = 0x25800000
	zipBytes[idx+24] = 0x00
	zipBytes[idx+25] = 0x00
	zipBytes[idx+26] = 0x80
	zipBytes[idx+27] = 0x25

	// Mock server to serve this zip file.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(zipBytes)))
		_, _ = w.Write(zipBytes)
	}))
	defer ts.Close()

	tempDir, err := os.MkdirTemp("", "ffmpeg_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()
	destPath := filepath.Join(tempDir, "ffmpeg")

	err = downloadAndExtractZip(context.Background(), ts.URL, destPath, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "zip entry too large") {
		t.Errorf("expected 'zip entry too large' error, got: %v", err)
	}
}

func TestDownloadAndExtractZip_LimitReader(t *testing.T) {
	// Test that io.LimitReader actually limits
	r := strings.NewReader("1234567890")
	lr := io.LimitReader(r, 5)
	content, _ := io.ReadAll(lr)
	if string(content) != "12345" {
		t.Errorf("expected 12345, got %s", string(content))
	}
}

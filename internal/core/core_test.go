package core_test

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go.massi.dev/webcamtimelapse/internal/core"
	"go.massi.dev/webcamtimelapse/internal/core/mocks"
)

type CoreTestSuite struct {
	suite.Suite
}

func (s *CoreTestSuite) SetupTest() {
}

func (s *CoreTestSuite) TestNewFrameContext_CreatesTempDir() {
	fc, err := core.NewFrameContext()
	s.Require().NoError(err)
	s.Require().NotNil(fc)

	s.Assert().DirExists(fc.TempDir)
	fc.Cleanup()
}

func (s *CoreTestSuite) TestFrameContext_Cleanup_RemovesDir() {
	fc, err := core.NewFrameContext()
	s.Require().NoError(err)

	dir := fc.TempDir
	fc.Cleanup()

	s.Assert().NoDirExists(dir)
}

func (s *CoreTestSuite) TestFrameContext_Cleanup_Idempotent() {
	fc, err := core.NewFrameContext()
	s.Require().NoError(err)

	s.Assert().NotPanics(func() {
		fc.Cleanup()
		fc.Cleanup() // second call must not panic
	})
}

func (s *CoreTestSuite) TestFetchAndSaveFrame_WritesJPEG() {
	client := mocks.NewHTTPClient(s.T())

	// Serve a minimal valid JPEG (1×1 white pixel).
	jpegData := minimalJPEG()

	client.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(jpegData)),
	}, nil)

	fc, err := core.NewFrameContext()
	s.Require().NoError(err)
	fc = fc.WithClient(client)
	defer fc.Cleanup()

	_, err = fc.FetchAndSaveFrame(context.Background(), "http://example.com/image.jpg", 0)
	s.Require().NoError(err)

	expected := filepath.Join(fc.TempDir, "frame_000000.jpg")
	info, err := os.Stat(expected)
	s.Require().NoError(err)
	s.Assert().Greater(info.Size(), int64(0))
}

func (s *CoreTestSuite) TestFetchAndSaveFrame_RequestError() {
	client := mocks.NewHTTPClient(s.T())

	client.On("Do", mock.Anything).Return((*http.Response)(nil), errors.New("request error"))

	fc, err := core.NewFrameContext()
	s.Require().NoError(err)
	fc = fc.WithClient(client)
	defer fc.Cleanup()

	_, err = fc.FetchAndSaveFrame(context.Background(), "http://example.com/image.jpg", 0)
	s.Assert().ErrorContains(err, "failed to fetch image: request error")
}

func (s *CoreTestSuite) TestFetchAndSaveFrame_BadStatus() {
	client := mocks.NewHTTPClient(s.T())

	client.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(bytes.NewReader([]byte{})),
	}, nil)

	fc, err := core.NewFrameContext()
	s.Require().NoError(err)
	fc = fc.WithClient(client)
	defer fc.Cleanup()

	_, err = fc.FetchAndSaveFrame(context.Background(), "http://example.com/image.jpg", 0)
	s.Assert().ErrorContains(err, "bad status")
}

func (s *CoreTestSuite) TestFetchAndSaveFrame_InvalidJPEG() {
	client := mocks.NewHTTPClient(s.T())

	client.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte("not a jpeg"))),
	}, nil)

	fc, err := core.NewFrameContext()
	s.Require().NoError(err)
	fc = fc.WithClient(client)
	defer fc.Cleanup()

	_, err = fc.FetchAndSaveFrame(context.Background(), "http://example.com/image.jpg", 0)
	s.Assert().ErrorContains(err, "failed to decode jpeg")
}

func (s *CoreTestSuite) TestFetchAndSaveFrame_ContextError() {
	client := mocks.NewHTTPClient(s.T())

	client.On("Do", mock.Anything).Return((*http.Response)(nil), context.Canceled)

	fc, err := core.NewFrameContext()
	s.Require().NoError(err)
	fc = fc.WithClient(client)
	defer fc.Cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel before request

	_, err = fc.FetchAndSaveFrame(ctx, "http://example.com/image.jpg", 0)
	s.Assert().ErrorContains(err, "failed to fetch image: context canceled")
}

func (s *CoreTestSuite) TestFetchAndSaveFrame_RequestCreationError() {
	// A mock client is not used because NewRequestWithContext fails first.
	fc, err := core.NewFrameContext()
	s.Require().NoError(err)
	defer fc.Cleanup()

	_, err = fc.FetchAndSaveFrame(context.Background(), "://invalid-url", 0) // Invalid URL causes NewRequest to fail
	s.Assert().ErrorContains(err, "failed to create request: parse")
}

func (s *CoreTestSuite) TestFetchAndSaveFrame_OutputFileCreationError() {
	client := mocks.NewHTTPClient(s.T())

	// Serve a minimal valid JPEG (1×1 white pixel).
	jpegData := minimalJPEG()

	client.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(jpegData)),
	}, nil)

	fc, err := core.NewFrameContext()
	s.Require().NoError(err)
	fc = fc.WithClient(client)
	defer fc.Cleanup()

	// Induce a file creation error by destroying the destination directory
	err = os.RemoveAll(fc.TempDir)
	s.Require().NoError(err)

	_, err = fc.FetchAndSaveFrame(context.Background(), "http://example.com/image.jpg", 0)
	s.Assert().ErrorContains(err, "failed to create output file: open")
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

func TestCoreTestSuite(t *testing.T) {
	suite.Run(t, new(CoreTestSuite))
}

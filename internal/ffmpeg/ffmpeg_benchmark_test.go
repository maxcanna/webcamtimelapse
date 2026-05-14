package ffmpeg

import (
	"context"
	"io"
	"testing"
)

type mockReader struct {
	size int64
	read int64
}

func (m *mockReader) Read(p []byte) (n int, err error) {
	if m.read >= m.size {
		return 0, io.EOF
	}
	n = len(p)
	if int64(n) > m.size-m.read {
		n = int(m.size - m.read)
	}
	m.read += int64(n)
	return n, nil
}

func BenchmarkPassThruReader_Read(b *testing.B) {
	ctx := context.Background()
	totalSize := int64(10 * 1024 * 1024) // 10 MB
	readSize := 1024                     // 1 KB reads to trigger many updates

	b.Run("Optimized", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			progress := make(chan SetupProgress, 1000) // Buffer to avoid blocking
			done := make(chan struct{})
			go func() {
				for range progress {
				}
				close(done)
			}()

			pt := &PassThruReader{
				Reader:   &mockReader{size: totalSize},
				ctx:      ctx,
				progress: progress,
				total:    totalSize,
				lastPct:  -1,
			}

			p := make([]byte, readSize)
			for {
				_, err := pt.Read(p)
				if err == io.EOF {
					break
				}
			}
			close(progress)
			<-done
		}
	})
}

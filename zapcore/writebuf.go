package zapcore

import (
	"bufio"
	"os"
)

// log with bufio
type bufWriterSync struct {
	buf        *bufio.Writer
	outputPath string
	close      func() error
	size       int
}

// Buffer wraps a WriteSyncer with bufio
func Buffer(f *os.File, size int, outputPath string) WriteSyncer {
	bw := &bufWriterSync{
		buf:        bufio.NewWriterSize(f, size),
		outputPath: outputPath,
		close:      f.Close,
		size:       size,
	}
	w := Lock(bw) // need lock for concurrence safe
	return w
}

func (w *bufWriterSync) Sync() error {
	return w.buf.Flush()
}

func (w *bufWriterSync) Write(p []byte) (written int, err error) {
	return w.buf.Write(p)
}

func (w *bufWriterSync) ReOpen() (err error) {
	w.Sync()
	w.close()
	f, err := os.OpenFile(w.outputPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		f.Close()
		return
	}
	w.buf = bufio.NewWriterSize(f, w.size)
	w.close = f.Close
	return nil
}

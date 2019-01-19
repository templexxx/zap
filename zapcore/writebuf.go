package zapcore

import (
	"bufio"
	"os"
	"time"
)

// log with bufio
type bufWriterSync struct {
	buf  *bufio.Writer
	size int

	outputPath string
	f          *os.File

	c chan *os.File
}

// Buffer wraps a WriteSyncer with bufio
func Buffer(f *os.File, size, flush int, outputPath string) WriteSyncer {
	bw := &bufWriterSync{
		buf:  bufio.NewWriterSize(f, size),
		size: size,

		outputPath: outputPath,
		f:          f,

		c: make(chan *os.File),
	}

	go cleanOldFile(bw.c)

	w := Lock(bw) // need lock for concurrence safe

	go func() {
		ticker := time.NewTicker(time.Duration(flush) * time.Second)
		for range ticker.C {
			w.Sync()
		}
	}()

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
	w.c <- w.f // non-blocking here for avoiding stuck all log write
	f, err := os.OpenFile(w.outputPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	w.buf = bufio.NewWriterSize(f, w.size)
	w.f = f
	return nil
}

// cleanOldFile will close, sync, drop page_cache
func cleanOldFile(c chan *os.File) {
	for f := range c {
		defer f.Close()
		info, err := f.Stat()
		if err != nil {
			continue
		}
		size := info.Size()

		f.Sync()
		dropCache(f, 0, size)
	}
}

const posix_fadv_dontneed = 4

// dropCache drop page_cache in range
func dropCache(f *os.File, offset, size int64) (err error) {

	return fadvise(f, offset, size, posix_fadv_dontneed)
}

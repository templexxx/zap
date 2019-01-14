package zapcore

import "os"

func fadvise(f *os.File, offset, size int64, advice int) (err error) {
	return
}

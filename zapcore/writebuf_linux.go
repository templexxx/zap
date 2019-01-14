package zapcore

import (
	"os"
	"syscall"
)

func fadvise(f *os.File, offset, size int64, advice int) (err error) {

	// discard partial pages are ignored
	var align int64
	align = 1 << 12
	size = (size + align - 1) &^ (align - 1)

	_, _, errno := syscall.Syscall6(syscall.SYS_FADVISE64, f.Fd(), uintptr(offset), uintptr(size), uintptr(advice), 0, 0)
	if errno != 0 {
		err = errno
	}
	return
}

package log

import (
	"os"
	"syscall"
)

func errorHandler(err error) {
	if e, ok := err.(*os.PathError); ok {
		err = e.Err
	}
	if err != syscall.EPIPE && err != syscall.ERROR_BROKEN_PIPE {
		return
	}
	os.Exit(5)
}

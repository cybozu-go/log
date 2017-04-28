// +build !windows

package log

import (
	"os"
	"syscall"
)

func errorHandler(err error) {
	if e, ok := err.(*os.PathError); ok {
		err = e.Err
	}
	if err != syscall.EPIPE {
		return
	}
	os.Exit(5)
}

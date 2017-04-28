package main

import (
	"flag"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/cybozu-go/log"
)

var (
	flgIgnoreSigPipe     = flag.Bool("i", false, "ignore SIGPIPE")
	flgClearErrorHandler = flag.Bool("c", false, "clear error handler")
	flgStdout            = flag.Bool("s", false, "output to stdout")
)

func main() {
	flag.Parse()
	if *flgIgnoreSigPipe {
		ch := make(chan os.Signal)
		signal.Notify(ch, syscall.SIGPIPE)
	}

	logger := log.NewLogger()
	if *flgClearErrorHandler {
		logger.SetErrorHandler(nil)
	}

	c := exec.Command("/bin/true")
	p, err := c.StdinPipe()
	if err != nil {
		log.ErrorExit(err)
	}
	err = c.Start()
	if err != nil {
		log.ErrorExit(err)
	}

	logger.SetOutput(p)
	if *flgStdout {
		logger.SetOutput(os.Stdout)
	}

	for {
		logger.Error("foo", nil)
		time.Sleep(time.Second)
	}
}
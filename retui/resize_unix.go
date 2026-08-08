//go:build !windows

package retui

import (
	"os"
	"os/signal"
	"syscall"
)

func newResizeChan() (chan os.Signal, func()) {
	resize := make(chan os.Signal, 1)
	signal.Notify(resize, syscall.SIGWINCH)
	return resize, func() { signal.Stop(resize) }
}

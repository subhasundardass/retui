//go:build windows

package retui

import (
	"os"
	"time"

	"golang.org/x/term"
)

// newResizeChan polls the console size on Windows (no SIGWINCH exists there)
// and pushes onto the channel whenever the size changes.
func newResizeChan() (chan os.Signal, func()) {
	resize := make(chan os.Signal, 1)
	done := make(chan struct{})

	go func() {
		lastW, lastH, _ := term.GetSize(int(os.Stdout.Fd()))
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				w, h, err := term.GetSize(int(os.Stdout.Fd()))
				if err != nil {
					continue
				}
				if w != lastW || h != lastH {
					lastW, lastH = w, h
					select {
					case resize <- os.Interrupt: // placeholder value, just used to wake the select loop
					default:
					}
				}
			}
		}
	}()

	return resize, func() { close(done) }
}

package debug

import (
	"fmt"
	rtdebug "runtime/debug"
)

// RecoverAndLog returns a function meant to be called directly with
// defer, e.g.:
//
//	defer debug.RecoverAndLog("render loop")()
//
// If a panic occurs, it is logged at LevelError (including a stack
// trace) through this package's logger and then swallowed, letting the
// caller (e.g. a RetUI application) restore terminal state and exit
// gracefully instead of crashing with a raw stack dump. Logging happens
// regardless of Enabled/SetLevel, so panics are never silently lost.
func RecoverAndLog(context string) func() {
	return func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("panic in %s: %v\n%s", context, r, rtdebug.Stack())
			writeEntry(LevelError, msg)
		}
	}
}

// RecoverAndReraise is like RecoverAndLog but re-panics with the original
// value after logging, for cases where a caller further up the stack
// still needs to handle the panic itself (e.g. a top-level main that
// restores the terminal and prints a crash report before exiting).
func RecoverAndReraise(context string) func() {
	return func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("panic in %s: %v\n%s", context, r, rtdebug.Stack())
			writeEntry(LevelError, msg)
			panic(r)
		}
	}
}

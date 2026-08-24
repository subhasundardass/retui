// Package debug provides an opt-in, near-zero-overhead diagnostics system
// for RetUI applications: leveled logging, memory statistics, frame/render
// timing, and panic recovery. Everything is disabled by default so that
// production builds pay effectively no runtime cost (a couple of atomic
// loads) until a caller explicitly turns it on.
//
// Typical use in an application's entrypoint:
//
//	debug.Enable()
//	debug.SetLevel(debug.LevelInfo)
//	defer debug.RecoverAndLog("main")()
//
//	debug.Infof("starting app, version=%s", version)
//
// All exported functions are safe for concurrent use.
package debug

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// Level controls the verbosity of the debug logger. Levels are ordered
// from least to most verbose; setting a level enables that level and every
// level above it (e.g. LevelDebug also emits Error/Warn/Info output).
type Level int32

const (
	// LevelOff disables all leveled log output. This is the default.
	LevelOff Level = iota
	LevelError
	LevelWarn
	LevelInfo
	LevelDebug
	LevelTrace
)

// String returns a human-readable name for the level, e.g. "INFO".
func (l Level) String() string {
	switch l {
	case LevelOff:
		return "OFF"
	case LevelError:
		return "ERROR"
	case LevelWarn:
		return "WARN"
	case LevelInfo:
		return "INFO"
	case LevelDebug:
		return "DEBUG"
	case LevelTrace:
		return "TRACE"
	default:
		return "UNKNOWN"
	}
}

const defaultRingCapacity = 512

var (
	enabledFlag int32 // atomic bool; 0 = disabled, 1 = enabled
	levelFlag   int32 = int32(LevelOff)

	outMu  sync.RWMutex
	output io.Writer = os.Stdout // matches the original PrintMemory's fmt.Printf behavior

	entries = newRingBuffer(defaultRingCapacity)

	// timeNow is indirected so tests can substitute it; production code
	// should never need to touch this.
	timeNow = time.Now
)

// Enable turns the debug system on. Enable and SetLevel are independent:
// Enable flips the master switch checked by every hot-path helper, and
// SetLevel controls which severities pass through once it's on.
func Enable() { atomic.StoreInt32(&enabledFlag, 1) }

// Disable turns the debug system off. Log*, PrintMemory's gated sibling
// LogMemory, and frame timing calls become a single atomic load once
// disabled. Note PrintMemory itself is unconditional; see its docs.
func Disable() { atomic.StoreInt32(&enabledFlag, 0) }

// Enabled reports whether the debug system is currently on.
func Enabled() bool { return atomic.LoadInt32(&enabledFlag) == 1 }

// SetLevel sets the minimum severity that will be logged.
func SetLevel(l Level) { atomic.StoreInt32(&levelFlag, int32(l)) }

// GetLevel returns the currently configured level.
func GetLevel() Level { return Level(atomic.LoadInt32(&levelFlag)) }

// SetOutput redirects where log lines and PrintMemory write. The default
// is os.Stdout. Passing nil restores the default.
func SetOutput(w io.Writer) {
	outMu.Lock()
	defer outMu.Unlock()
	if w == nil {
		w = os.Stdout
	}
	output = w
}

func currentOutput() io.Writer {
	outMu.RLock()
	defer outMu.RUnlock()
	return output
}

// SetHistorySize resizes the retained in-memory log ring buffer used by
// Entries. The default capacity is 512 entries. Existing entries are
// preserved (most recent first) up to the new capacity.
func SetHistorySize(n int) { entries.resize(n) }

// Entries returns a snapshot of retained log entries, oldest first,
// independent of the current output writer. Useful for an in-app debug
// overlay, or for dumping recent history from a panic handler.
func Entries() []Entry { return entries.snapshot() }

// ClearEntries discards all retained log history.
func ClearEntries() { entries.clear() }

// shouldLog is the hot-path gate. It costs two atomic loads when the
// system is disabled and never allocates.
func shouldLog(l Level) bool {
	if atomic.LoadInt32(&enabledFlag) == 0 {
		return false
	}
	return l <= Level(atomic.LoadInt32(&levelFlag))
}

func logf(l Level, format string, args ...interface{}) {
	if !shouldLog(l) {
		return
	}
	msg := fmt.Sprintf(format, args...)
	writeEntry(l, msg)
}

func writeEntry(l Level, msg string) {
	t := timeNow()
	entries.push(Entry{Time: t, Level: l, Message: msg})
	fmt.Fprintf(currentOutput(), "%s [%s] %s\n", t.Format("15:04:05.000"), l, msg)
}

// Errorf logs a message at LevelError.
func Errorf(format string, args ...interface{}) { logf(LevelError, format, args...) }

// Warnf logs a message at LevelWarn.
func Warnf(format string, args ...interface{}) { logf(LevelWarn, format, args...) }

// Infof logs a message at LevelInfo.
func Infof(format string, args ...interface{}) { logf(LevelInfo, format, args...) }

// Debugf logs a message at LevelDebug.
func Debugf(format string, args ...interface{}) { logf(LevelDebug, format, args...) }

// Tracef logs a message at LevelTrace, the most verbose level. Intended
// for per-frame or per-event diagnostics; expect this to be noisy.
func Tracef(format string, args ...interface{}) { logf(LevelTrace, format, args...) }

// Fatalf always logs a message at LevelError severity, bypassing
// Enable/SetLevel gating entirely, and is still captured in Entries like
// any other log call. It mirrors the unconditional path RecoverAndLog
// already uses for panics, and exists for crash-adjacent messages that
// must never be silently dropped just because the debug system happens
// to be off or filtered above Error.
//
// Fatalf does not terminate the process — callers that need that (e.g.
// retui.Fatal) call os.Exit themselves after logging.
func Fatalf(format string, args ...interface{}) {
	writeEntry(LevelError, fmt.Sprintf(format, args...))
}

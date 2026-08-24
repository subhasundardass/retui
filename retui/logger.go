package retui

//--------USE ANYWHERE-
// retui.Debug("Render Dashboard")
// retui.Debug("Focused:", focused)
// retui.Debug("Current Route:", route)
//----------
//
// Internally this file is now a thin, backward-compatible facade over
// github.com/subhasundardass/retui/debug — the actual leveled logger,
// in-memory history ring buffer, and file output all live there. Every
// exported name and signature below is unchanged from the original
// implementation; only what's underneath moved.
import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/subhasundardass/retui/retui/debug"
)

var (
	debugMode bool
	logFile   *os.File
	logMu     sync.RWMutex
	logLevel  LogLevel
)

// LogLevel represents the logging level.
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelSuccess
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String returns the string representation of the log level.
func (l LogLevel) String() string {
	switch l {
	case LevelSuccess:
		return "SUCCESS"
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// toDebugLevel maps retui's LogLevel onto debug.Level. LevelSuccess has
// no dedicated debug.Level — Success/Successf tag it via a "[SUCCESS]"
// message prefix instead — so it maps to debug.LevelInfo for gating
// purposes. LevelFatal is handled separately (see Fatal/Fatalf) via
// debug.Fatalf's unconditional path, since a fatal message must never be
// silently dropped by level filtering right before the process exits.
func toDebugLevel(l LogLevel) debug.Level {
	switch l {
	case LevelDebug:
		return debug.LevelDebug
	case LevelSuccess, LevelInfo:
		return debug.LevelInfo
	case LevelWarn:
		return debug.LevelWarn
	case LevelError, LevelFatal:
		return debug.LevelError
	default:
		return debug.LevelInfo
	}
}

func init() {
	debugMode = true
	logLevel = LevelDebug

	// This package's original design was always-on (init() enabled
	// debugMode and opened the log file unconditionally); SetDebugMode
	// and SetLogLevel only ever adjusted verbosity, never a master
	// switch. debug.Enable() is called once here, permanently, to
	// preserve that: SetDebugMode/SetLogLevel below adjust debug.SetLevel
	// only.
	debug.Enable()
	debug.SetLevel(toDebugLevel(logLevel))

	if err := setupLogging(); err != nil {
		// Fall back to stderr if file logging fails, same as before.
		debug.SetOutput(os.Stderr)
		debug.Errorf("Failed to setup log file: %v", err)
	}
}

// setupLogging configures the logging system.
func setupLogging() error {
	logMu.Lock()
	defer logMu.Unlock()

	logDir := getLogDir()
	if err := os.MkdirAll(logDir, 0750); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// logPath is built entirely from getLogDir() (cwd) and a fixed
	// filename — never from user input or external config — so path
	// traversal isn't possible here despite gosec flagging the variable.
	logPath := filepath.Join(logDir, "retui.log")
	var err error
	logFile, err = os.OpenFile(
		logPath, // #nosec G304 -- logPath is derived from getLogDir()+fixed filename, not external input
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0600,
	)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// File only — no stdout/stderr, so the TUI screen stays clean.
	debug.SetOutput(logFile)

	return nil
}

// getLogDir returns the directory for log files.
func getLogDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

// SetDebugMode enables or disables debug mode.
func SetDebugMode(enabled bool) {
	logMu.Lock()
	debugMode = enabled
	if enabled {
		logLevel = LevelDebug
	} else {
		logLevel = LevelInfo
	}
	l := logLevel
	logMu.Unlock()

	debug.SetLevel(toDebugLevel(l))
}

// SetLogLevel sets the minimum log level.
func SetLogLevel(level LogLevel) {
	logMu.Lock()
	logLevel = level
	logMu.Unlock()

	debug.SetLevel(toDebugLevel(level))
}

// GetLogLevel returns the current log level.
func GetLogLevel() LogLevel {
	logMu.RLock()
	defer logMu.RUnlock()
	return logLevel
}

// IsDebugMode returns true if debug mode is enabled.
func IsDebugMode() bool {
	logMu.RLock()
	defer logMu.RUnlock()
	return debugMode
}

// CloseLog closes the log file.
func CloseLog() error {
	logMu.Lock()
	defer logMu.Unlock()

	if logFile != nil {
		return logFile.Close()
	}
	return nil
}

// callerInfo replicates the "file:line" format the original logger
// embedded in every line. skip is the runtime.Caller depth from the
// exported function that calls this directly (2 for a top-level
// function like Debug/Info/etc. calling callerInfo itself).
func callerInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	parts := strings.Split(file, "/")
	return fmt.Sprintf("%s:%d", parts[len(parts)-1], line)
}

// Debug logs debug messages.
//
// Note: the original implementation additionally gated this behind a
// direct `if debugMode` check, on top of the level comparison SetLogLevel
// already performs. That check was redundant whenever SetDebugMode was
// used (it sets logLevel in lockstep with debugMode) and only mattered
// if SetLogLevel(LevelDebug) was called directly without SetDebugMode —
// in which case it silently suppressed Debug/Debugf despite the level
// otherwise permitting it. That inconsistency is removed here: Debug and
// Debugf are now gated purely by the configured level, like every other
// call in this file.
func Debug(args ...interface{}) {
	debug.Debugf("%s %s", callerInfo(2), fmt.Sprint(args...))
}

// Debugf logs a formatted debug message. See the Debug doc comment for a
// note on a small behavior fix relative to the original implementation.
func Debugf(format string, args ...interface{}) {
	debug.Debugf("%s %s", callerInfo(2), fmt.Sprintf(format, args...))
}

// Info logs info messages.
func Info(args ...interface{}) {
	debug.Infof("%s %s", callerInfo(2), fmt.Sprint(args...))
}

// Infof logs a formatted info message.
func Infof(format string, args ...interface{}) {
	debug.Infof("%s %s", callerInfo(2), fmt.Sprintf(format, args...))
}

// Warn logs warning messages.
func Warn(args ...interface{}) {
	debug.Warnf("%s %s", callerInfo(2), fmt.Sprint(args...))
}

// Warnf logs a formatted warning message.
func Warnf(format string, args ...interface{}) {
	debug.Warnf("%s %s", callerInfo(2), fmt.Sprintf(format, args...))
}

// Error logs error messages.
func Error(args ...interface{}) {
	debug.Errorf("%s %s", callerInfo(2), fmt.Sprint(args...))
}

// Errorf logs a formatted error message.
func Errorf(format string, args ...interface{}) {
	debug.Errorf("%s %s", callerInfo(2), fmt.Sprintf(format, args...))
}

// Fatal logs a fatal message and exits. Logging is unconditional — via
// debug.Fatalf — regardless of SetDebugMode/SetLogLevel, because a fatal
// message must never be silently dropped by verbosity filtering right
// before the process exits.
func Fatal(args ...interface{}) {
	debug.Fatalf("%s %s", callerInfo(2), fmt.Sprint(args...))
	os.Exit(1)
}

// Fatalf logs a formatted fatal message and exits. See Fatal.
func Fatalf(format string, args ...interface{}) {
	debug.Fatalf("%s %s", callerInfo(2), fmt.Sprintf(format, args...))
	os.Exit(1)
}

// Success logs a success message, tagged "[SUCCESS]" in the message body
// (debug.Level has no dedicated success level, so the tag lives in the
// text rather than the bracketed level prefix).
func Success(args ...interface{}) {
	debug.Infof("%s [SUCCESS] \u2705 %s", callerInfo(2), fmt.Sprint(args...))
}

// Successf logs a formatted success message. See Success.
//
// Note: the original Success() logged at LevelInfo while Successf()
// logged at LevelSuccess — two different severities for what's
// conceptually the same kind of message. Both now behave identically.
func Successf(format string, args ...interface{}) {
	debug.Infof("%s [SUCCESS] \u2705 %s", callerInfo(2), fmt.Sprintf(format, args...))
}

// LogWithFields logs with additional structured fields.
func LogWithFields(level LogLevel, fields map[string]interface{}, message string) {
	fieldStr := ""
	if len(fields) > 0 {
		fieldStr = " " + fmt.Sprint(fields)
	}
	full := message + fieldStr

	if level == LevelFatal {
		debug.Fatalf("%s %s", callerInfo(2), full)
		return
	}

	switch toDebugLevel(level) {
	case debug.LevelDebug:
		debug.Debugf("%s %s", callerInfo(2), full)
	case debug.LevelWarn:
		debug.Warnf("%s %s", callerInfo(2), full)
	case debug.LevelError:
		debug.Errorf("%s %s", callerInfo(2), full)
	default:
		debug.Infof("%s %s", callerInfo(2), full)
	}
}

// Logger is a logger bound to a fixed set of structured fields.
type Logger struct {
	fields map[string]interface{}
}

// NewLogger creates a new logger with fields.
func NewLogger(fields map[string]interface{}) *Logger {
	return &Logger{fields: fields}
}

// Debug logs a debug message with fields.
func (l *Logger) Debug(message string) {
	LogWithFields(LevelDebug, l.fields, message)
}

// Info logs an info message with fields.
func (l *Logger) Info(message string) {
	LogWithFields(LevelInfo, l.fields, message)
}

// Warn logs a warning message with fields.
func (l *Logger) Warn(message string) {
	LogWithFields(LevelWarn, l.fields, message)
}

// Error logs an error message with fields.
func (l *Logger) Error(message string) {
	LogWithFields(LevelError, l.fields, message)
}

// AddField adds a field to the logger.
func (l *Logger) AddField(key string, value interface{}) *Logger {
	if l.fields == nil {
		l.fields = make(map[string]interface{})
	}
	l.fields[key] = value
	return l
}

// WithFields creates a new logger with additional fields.
func WithFields(fields map[string]interface{}) *Logger {
	return NewLogger(fields)
}

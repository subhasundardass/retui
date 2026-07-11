package retui

//--------USE ANYWHERE-
// retui.Debug("Render Dashboard")
// retui.Debug("Focused:", focused)
// retui.Debug("Current Route:", route)
//----------
import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	logger     *log.Logger
	debugMode  bool
	logFile    *os.File
	logMu      sync.RWMutex
	logLevel   LogLevel
	logOutputs []io.Writer
)

// LogLevel represents the logging level
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelSuccess
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String returns the string representation of the log level
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

// Color codes for terminal output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

func init() {
	// Initialize with default settings
	debugMode = true
	logLevel = LevelDebug
	logOutputs = make([]io.Writer, 0)

	// Setup logging
	if err := setupLogging(); err != nil {
		// Fallback to stderr if file logging fails
		logger = log.New(os.Stderr, "", log.LstdFlags)
		logger.Printf("Failed to setup log file: %v", err)
	}
}

// setupLogging configures the logging system
func setupLogging() error {
	logMu.Lock()
	defer logMu.Unlock()

	logDir := getLogDir()
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath := filepath.Join(logDir, "retui.log")
	var err error
	logFile, err = os.OpenFile(
		logPath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// File only — no stderr so TUI screen stays clean
	logger = log.New(logFile, "", 0)

	return nil
}

// getLogDir returns the directory for log files
func getLogDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		// Go up two levels to project root
		return filepath.Join(filepath.Dir(filename), "..")
	}
	return "./"
}

// SetDebugMode enables or disables debug mode
func SetDebugMode(enabled bool) {
	logMu.Lock()
	defer logMu.Unlock()
	debugMode = enabled
	if enabled {
		logLevel = LevelDebug
	} else {
		logLevel = LevelInfo
	}
}

// SetLogLevel sets the minimum log level
func SetLogLevel(level LogLevel) {
	logMu.Lock()
	defer logMu.Unlock()
	logLevel = level
}

// GetLogLevel returns the current log level
func GetLogLevel() LogLevel {
	logMu.RLock()
	defer logMu.RUnlock()
	return logLevel
}

// IsDebugMode returns true if debug mode is enabled
func IsDebugMode() bool {
	logMu.RLock()
	defer logMu.RUnlock()
	return debugMode
}

// CloseLog closes the log file
func CloseLog() error {
	logMu.Lock()
	defer logMu.Unlock()

	if logFile != nil {
		return logFile.Close()
	}
	return nil
}

// logMessage is the internal logging function
func logMessage(level LogLevel, color string, args ...interface{}) {
	logMu.RLock()
	defer logMu.RUnlock()

	// Check if this level should be logged
	if level < logLevel {
		return
	}

	if logger == nil {
		return
	}

	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	callerInfo := ""
	if ok {
		// Get only the filename, not the full path
		parts := strings.Split(file, "/")
		filename := parts[len(parts)-1]
		callerInfo = fmt.Sprintf("%s:%d", filename, line)
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	levelStr := level.String()

	// Format the message
	message := fmt.Sprint(args...)

	// Create log entry
	logEntry := fmt.Sprintf("[%s] [%s] %s %s",
		timestamp,
		levelStr,
		callerInfo,
		message,
	)

	// Write to log
	logger.Println(logEntry)

	// Also print to stdout with color if it's a terminal
	// if isTerminal() {
	// 	fmt.Printf("%s%s%s\n", color, logEntry, colorReset)
	// }
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// Debug logs debug messages (only shown in debug mode)
func Debug(args ...interface{}) {
	if debugMode {
		logMessage(LevelDebug, colorCyan, args...)
	}
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	if debugMode {
		logMessage(LevelDebug, colorCyan, fmt.Sprintf(format, args...))
	}
}

// Info logs info messages
func Info(args ...interface{}) {
	logMessage(LevelInfo, colorGreen, args...)
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	logMessage(LevelInfo, colorGreen, fmt.Sprintf(format, args...))
}

// Warn logs warning messages
func Warn(args ...interface{}) {
	logMessage(LevelWarn, colorYellow, args...)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	logMessage(LevelWarn, colorYellow, fmt.Sprintf(format, args...))
}

// Error logs error messages
func Error(args ...interface{}) {
	logMessage(LevelError, colorRed, args...)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	logMessage(LevelError, colorRed, fmt.Sprintf(format, args...))
}

// Fatal logs fatal messages and exits
func Fatal(args ...interface{}) {
	logMessage(LevelFatal, colorRed, args...)
	os.Exit(1)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, args ...interface{}) {
	logMessage(LevelFatal, colorRed, fmt.Sprintf(format, args...))
	os.Exit(1)
}

// Success logs success messages (green)
func Success(args ...interface{}) {
	logMessage(LevelInfo, colorGreen, "✅ "+fmt.Sprint(args...))
}

// Successf logs a formatted success message
func Successf(format string, args ...interface{}) {
	logMessage(LevelSuccess, colorBlue, "✅ "+fmt.Sprintf(format, args...))
}

// LogWithFields logs with additional fields (like structured logging)
func LogWithFields(level LogLevel, fields map[string]interface{}, message string) {
	if level < logLevel {
		return
	}

	fieldStr := ""
	if len(fields) > 0 {
		fieldStr = " " + fmt.Sprint(fields)
	}

	logMessage(level, colorWhite, message+fieldStr)
}

// WithField returns a new logger with a field (for structured logging)
type Logger struct {
	fields map[string]interface{}
}

// NewLogger creates a new logger with fields
func NewLogger(fields map[string]interface{}) *Logger {
	return &Logger{fields: fields}
}

// Debug logs a debug message with fields
func (l *Logger) Debug(message string) {
	if debugMode {
		LogWithFields(LevelDebug, l.fields, message)
	}
}

// Info logs an info message with fields
func (l *Logger) Info(message string) {
	LogWithFields(LevelInfo, l.fields, message)
}

// Warn logs a warning message with fields
func (l *Logger) Warn(message string) {
	LogWithFields(LevelWarn, l.fields, message)
}

// Error logs an error message with fields
func (l *Logger) Error(message string) {
	LogWithFields(LevelError, l.fields, message)
}

// AddField adds a field to the logger
func (l *Logger) AddField(key string, value interface{}) *Logger {
	if l.fields == nil {
		l.fields = make(map[string]interface{})
	}
	l.fields[key] = value
	return l
}

// WithFields creates a new logger with additional fields
func WithFields(fields map[string]interface{}) *Logger {
	return NewLogger(fields)
}

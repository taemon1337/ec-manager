package logger

import (
	"io"
	"log/slog"
	"os"
	"sync"
)

// LogLevel represents the logging level
type LogLevel string

const (
	// DebugLevel for detailed debugging information
	DebugLevel LogLevel = "debug"
	// InfoLevel for general operational information
	InfoLevel LogLevel = "info"
	// WarnLevel for warning messages
	WarnLevel LogLevel = "warn"
	// ErrorLevel for error messages
	ErrorLevel LogLevel = "error"
)

// Logger wraps slog.Logger to provide a consistent interface
type Logger struct {
	*slog.Logger
}

var (
	defaultLogger *Logger
	loggerOnce   sync.Once
	mu           sync.RWMutex
)

// NewLogger creates a new logger with the specified level and writer
func NewLogger(level LogLevel, w io.Writer) *Logger {
	var logLevel slog.Level
	switch level {
	case DebugLevel:
		logLevel = slog.LevelDebug
	case InfoLevel:
		logLevel = slog.LevelInfo
	case WarnLevel:
		logLevel = slog.LevelWarn
	case ErrorLevel:
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	handler := slog.NewTextHandler(w, opts)
	return &Logger{slog.New(handler)}
}

// Init initializes the default logger with the specified level
func Init(level LogLevel) {
	loggerOnce.Do(func() {
		defaultLogger = NewLogger(level, os.Stdout)
	})
}

// Get returns the default logger, initializing it if necessary
func Get() *Logger {
	mu.RLock()
	if defaultLogger == nil {
		mu.RUnlock()
		Init(InfoLevel)
		mu.RLock()
	}
	defer mu.RUnlock()
	return defaultLogger
}

// Debug logs at debug level
func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}

// Info logs at info level
func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

// Warn logs at warn level
func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

// Error logs at error level
func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}

// With returns a new logger with the given attributes
func With(args ...any) *slog.Logger {
	return Get().With(args...)
}

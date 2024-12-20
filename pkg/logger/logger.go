package logger

import (
	"io"
	"log/slog"
	"os"
	"sync"
)

var (
	logger     *slog.Logger
	loggerOnce sync.Once
	mu         sync.RWMutex
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

// Init initializes the logger with the specified level
func Init(level LogLevel) {
	loggerOnce.Do(func() {
		initLogger(level, os.Stdout)
	})
}

// InitWithWriter initializes the logger with a specific writer (useful for testing)
func InitWithWriter(level LogLevel, w io.Writer) {
	loggerOnce.Do(func() {
		initLogger(level, w)
	})
}

// Reset resets the logger state (useful for testing)
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	logger = nil
	loggerOnce = sync.Once{}
}

func initLogger(level LogLevel, w io.Writer) {
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
	l := slog.New(handler)

	mu.Lock()
	logger = l
	mu.Unlock()
}

func getLogger() *slog.Logger {
	mu.RLock()
	defer mu.RUnlock()
	if logger == nil {
		mu.RUnlock()
		Init(InfoLevel)
		mu.RLock()
	}
	return logger
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	getLogger().Debug(msg, args...)
}

// Info logs an info message
func Info(msg string, args ...any) {
	getLogger().Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	getLogger().Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	getLogger().Error(msg, args...)
}

// With returns a new logger with the given attributes
func With(args ...any) *slog.Logger {
	return getLogger().With(args...)
}

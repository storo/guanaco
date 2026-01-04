// Package logger provides logging functionality for Guanaco.
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/storo/guanaco/internal/config"
)

// Level represents the log level.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger handles application logging.
type Logger struct {
	mu       sync.Mutex
	level    Level
	file     *os.File
	logger   *log.Logger
	toStderr bool
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// Init initializes the default logger.
func Init() error {
	var initErr error
	once.Do(func() {
		defaultLogger, initErr = newLogger()
	})
	return initErr
}

func newLogger() (*Logger, error) {
	// Create log directory
	logDir := filepath.Join(config.GetDataDir(), "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file with date
	logFile := filepath.Join(logDir, fmt.Sprintf("guanaco_%s.log", time.Now().Format("2006-01-02")))
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Write to both file and stderr
	multiWriter := io.MultiWriter(file, os.Stderr)

	l := &Logger{
		level:    LevelInfo,
		file:     file,
		logger:   log.New(multiWriter, "", 0),
		toStderr: true,
	}

	// Check for debug mode
	if os.Getenv("GUANACO_DEBUG") == "1" {
		l.level = LevelDebug
	}

	l.Info("Logger initialized", "file", logFile)

	return l, nil
}

// Close closes the log file.
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// SetLevel sets the minimum log level.
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *Logger) log(level Level, msg string, keyvals ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	logLine := fmt.Sprintf("[%s] %s: %s", timestamp, level.String(), msg)

	// Add key-value pairs
	for i := 0; i < len(keyvals)-1; i += 2 {
		logLine += fmt.Sprintf(" %v=%v", keyvals[i], keyvals[i+1])
	}

	l.logger.Println(logLine)
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, keyvals ...interface{}) {
	l.log(LevelDebug, msg, keyvals...)
}

// Info logs an info message.
func (l *Logger) Info(msg string, keyvals ...interface{}) {
	l.log(LevelInfo, msg, keyvals...)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, keyvals ...interface{}) {
	l.log(LevelWarn, msg, keyvals...)
}

// Error logs an error message.
func (l *Logger) Error(msg string, keyvals ...interface{}) {
	l.log(LevelError, msg, keyvals...)
}

// Package-level functions that use the default logger

// Debug logs a debug message.
func Debug(msg string, keyvals ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Debug(msg, keyvals...)
	}
}

// Info logs an info message.
func Info(msg string, keyvals ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Info(msg, keyvals...)
	}
}

// Warn logs a warning message.
func Warn(msg string, keyvals ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Warn(msg, keyvals...)
	}
}

// Error logs an error message.
func Error(msg string, keyvals ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Error(msg, keyvals...)
	}
}

// Close closes the default logger.
func Close() error {
	if defaultLogger != nil {
		return defaultLogger.Close()
	}
	return nil
}

// LogFile returns the current log file path.
func LogFile() string {
	logDir := filepath.Join(config.GetDataDir(), "logs")
	return filepath.Join(logDir, fmt.Sprintf("guanaco_%s.log", time.Now().Format("2006-01-02")))
}

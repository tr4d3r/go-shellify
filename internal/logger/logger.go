package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

// Level represents the logging level
type Level int

const (
	// LevelDebug is the most verbose level
	LevelDebug Level = iota
	// LevelInfo is for informational messages
	LevelInfo
	// LevelWarn is for warning messages
	LevelWarn
	// LevelError is for error messages
	LevelError
	// LevelFatal is for fatal errors
	LevelFatal
)

// Logger represents a logger instance
type Logger struct {
	level      Level
	output     io.Writer
	prefix     string
	timeFormat string
	colors     bool
}

var (
	// Default is the default logger instance
	Default *Logger
	
	// Color codes for terminal output
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

func init() {
	// Initialize default logger
	Default = New(LevelInfo, os.Stdout)
	Default.SetColors(true)
}

// New creates a new logger instance
func New(level Level, output io.Writer) *Logger {
	return &Logger{
		level:      level,
		output:     output,
		timeFormat: "2006-01-02 15:04:05",
		colors:     false,
	}
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level Level) {
	l.level = level
}

// SetVerbose enables verbose (debug) logging
func (l *Logger) SetVerbose(verbose bool) {
	if verbose {
		l.level = LevelDebug
	} else {
		l.level = LevelInfo
	}
}

// SetOutput sets the output writer
func (l *Logger) SetOutput(output io.Writer) {
	l.output = output
}

// SetPrefix sets the logger prefix
func (l *Logger) SetPrefix(prefix string) {
	l.prefix = prefix
}

// SetColors enables or disables colored output
func (l *Logger) SetColors(enabled bool) {
	l.colors = enabled
}

// formatMessage formats a log message
func (l *Logger) formatMessage(level Level, format string, args ...interface{}) string {
	var levelStr, color string
	
	switch level {
	case LevelDebug:
		levelStr = "DEBUG"
		color = colorGray
	case LevelInfo:
		levelStr = "INFO"
		color = colorBlue
	case LevelWarn:
		levelStr = "WARN"
		color = colorYellow
	case LevelError:
		levelStr = "ERROR"
		color = colorRed
	case LevelFatal:
		levelStr = "FATAL"
		color = colorRed + colorBold
	}
	
	message := fmt.Sprintf(format, args...)
	
	// Build the log message
	var output strings.Builder
	
	// Add timestamp if not in simple mode
	if l.level == LevelDebug {
		timestamp := time.Now().Format(l.timeFormat)
		output.WriteString(timestamp)
		output.WriteString(" ")
	}
	
	// Add level
	if l.colors && color != "" {
		output.WriteString(color)
		output.WriteString("[")
		output.WriteString(levelStr)
		output.WriteString("]")
		output.WriteString(colorReset)
	} else {
		output.WriteString("[")
		output.WriteString(levelStr)
		output.WriteString("]")
	}
	output.WriteString(" ")
	
	// Add prefix if set
	if l.prefix != "" {
		output.WriteString(l.prefix)
		output.WriteString(": ")
	}
	
	// Add message
	output.WriteString(message)
	
	return output.String()
}

// log writes a log message if the level is enabled
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level >= l.level {
		message := l.formatMessage(level, format, args...)
		fmt.Fprintln(l.output, message)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(LevelFatal, format, args...)
	os.Exit(1)
}

// Success logs a success message (info level with green color)
func (l *Logger) Success(format string, args ...interface{}) {
	if LevelInfo >= l.level {
		message := fmt.Sprintf(format, args...)
		if l.colors {
			fmt.Fprintf(l.output, "\033[32m✓\033[0m %s\n", message)
		} else {
			fmt.Fprintf(l.output, "✓ %s\n", message)
		}
	}
}

// Package-level functions that use the default logger

// SetLevel sets the default logger's level
func SetLevel(level Level) {
	Default.SetLevel(level)
}

// SetVerbose enables verbose logging on the default logger
func SetVerbose(verbose bool) {
	Default.SetVerbose(verbose)
}

// Debug logs a debug message using the default logger
func Debug(format string, args ...interface{}) {
	Default.Debug(format, args...)
}

// Info logs an info message using the default logger
func Info(format string, args ...interface{}) {
	Default.Info(format, args...)
}

// Warn logs a warning message using the default logger
func Warn(format string, args ...interface{}) {
	Default.Warn(format, args...)
}

// Error logs an error message using the default logger
func Error(format string, args ...interface{}) {
	Default.Error(format, args...)
}

// Fatal logs a fatal message using the default logger and exits
func Fatal(format string, args ...interface{}) {
	Default.Fatal(format, args...)
}

// Success logs a success message using the default logger
func Success(format string, args ...interface{}) {
	Default.Success(format, args...)
}

// ParseLevel parses a string level to a Level
func ParseLevel(levelStr string) (Level, error) {
	switch strings.ToLower(levelStr) {
	case "debug":
		return LevelDebug, nil
	case "info":
		return LevelInfo, nil
	case "warn", "warning":
		return LevelWarn, nil
	case "error":
		return LevelError, nil
	case "fatal":
		return LevelFatal, nil
	default:
		return LevelInfo, fmt.Errorf("unknown log level: %s", levelStr)
	}
}

// StandardLogger returns a standard log.Logger that writes to this logger
func (l *Logger) StandardLogger() *log.Logger {
	return log.New(l.output, "", 0)
}
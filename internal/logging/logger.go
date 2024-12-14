package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Logger is a wrapper around slog.Logger
var logger *slog.Logger

// Define a custom FATAL level
const LevelFatal slog.Level = slog.Level(12)

// ExitFunc is a custom exit function that allows for testing of the FATAL level
var ExitFunc = os.Exit

// ANSI color codes
const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
)

// CustomTextHandler is a custom handler for slog that formats logs as specified
type CustomTextHandler struct {
	slog.Handler
	writer io.Writer
	level  slog.Level
}

// NewCustomTextHandler creates a new custom text handler
func NewCustomTextHandler(w io.Writer, opts *slog.HandlerOptions) *CustomTextHandler {
	textHandler := slog.NewTextHandler(w, opts)

	return &CustomTextHandler{
		Handler: textHandler,
		writer:  w,
		level:   opts.Level.(slog.Level),
	}
}

// Handle implements slog.Handler
func (h *CustomTextHandler) Handle(ctx context.Context, r slog.Record) error {

	// Set formats for output
	timestamp := r.Time.Format("2006/01/02 15:04:05")
	level := strings.ToUpper(r.Level.String())
	msg := r.Message

	var color string
	switch r.Level {
	case slog.LevelDebug:
		color = Blue
	case slog.LevelInfo:
		color = Green
	case slog.LevelWarn:
		color = Yellow
	case slog.LevelError:
		color = Red
	case LevelFatal:
		color = Magenta
	default:
		color = White
	}

	// Custom level mapping for FATAL
	if r.Level == LevelFatal {
		level = "FATAL"
	}

	// Write formatted output for all other logging levels
	fmt.Fprintf(h.writer, "%s %s%s%s %s%s\n", timestamp, color, level, Reset, msg, Reset)
	return nil
}

// Enabled implements slog.Handler
func (h *CustomTextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Handler.Enabled(ctx, level)
}

// WithAttrs implements slog.Handler which is used to add attributes
func (h *CustomTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CustomTextHandler{
		Handler: h.Handler.WithAttrs(attrs),
		writer:  h.writer,
	}
}

// WithGroup implements slog.Handler which is used to group logs
func (h *CustomTextHandler) WithGroup(name string) slog.Handler {
	return &CustomTextHandler{
		Handler: h.Handler.WithGroup(name),
		writer:  h.writer,
	}
}

// Initialize initializes the logger with a specified log level
func Initialize(logLevel string) *slog.Logger {

	// Set log level
	var level slog.Level

	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Create a new logger with a custom text handler
	logger = slog.New(NewCustomTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	return logger
}

// Info logs an informational message
func Info(msg string, keysAndValues ...interface{}) {
	logger.Info(msg, keysAndValues...)
}

// Warn logs a warning message
func Warn(msg string, keysAndValues ...interface{}) {
	logger.Warn(msg, keysAndValues...)
}

// Error logs an error message
func Error(msg string, keysAndValues ...interface{}) {
	logger.Error(msg, keysAndValues...)
}

// Debug logs a debug message
func Debug(msg string, keysAndValues ...interface{}) {
	logger.Debug(msg, keysAndValues...)
}

func Fatal(msg string, keysAndValues ...interface{}) {
	logger.Log(context.Background(), LevelFatal, msg, keysAndValues...)
	ExitFunc(1)
}

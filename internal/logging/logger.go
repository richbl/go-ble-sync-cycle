package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

var logger *slog.Logger

// ComponentType represents the type of component
type ComponentType string

const (
	APP   ComponentType = "[APP]"
	BLE   ComponentType = "[BLE]"
	SPEED ComponentType = "[SPEED]"
	VIDEO ComponentType = "[VIDEO]"
)

// Create a new slog level for the Fatal logging level
const LevelFatal slog.Level = slog.Level(12)

// ExitFunc represents the exit function (needed for testing)
var ExitFunc = os.Exit

// Color constants
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

// CustomTextHandler represents a custom text handler
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

func (h *CustomTextHandler) Handle(ctx context.Context, r slog.Record) error {

	// Create custom logger output
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

	if r.Level == LevelFatal {
		level = "FATAL"
	}

	var component string

	// Get component from attributes
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "component" {
			component = a.Value.String()
		}
		return true
	})

	// Write output format to writer
	msgPattern := "%s %s%s%s%s %s%s\n"

	if len(component) > 0 {
		msgPattern = "%s %s%s%s %s %s%s\n"
	}

	fmt.Fprintf(h.writer, msgPattern, timestamp, color, level, Reset, component, msg, Reset)
	return nil
}

// Enabled checks if the handler is enabled
func (h *CustomTextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Handler.Enabled(ctx, level)
}

// WithAttrs adds attributes to the handler
func (h *CustomTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CustomTextHandler{
		Handler: h.Handler.WithAttrs(attrs),
		writer:  h.writer,
	}
}

// WithGroup adds a group to the handler
func (h *CustomTextHandler) WithGroup(name string) slog.Handler {
	return &CustomTextHandler{
		Handler: h.Handler.WithGroup(name),
		writer:  h.writer,
	}
}

// Initialize sets up the logger
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

	// Initialize logger
	logger = slog.New(NewCustomTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	return logger
}

// logWithOptionalComponent logs a message with an optional component
func logWithOptionalComponent(level slog.Level, first interface{}, args ...interface{}) {
	var msg string
	var component string

	// Check if first argument is a ComponentType
	if comp, ok := first.(ComponentType); ok {
		if len(args) > 0 {
			msg = fmt.Sprint(args[0])
		}
		component = string(comp)
	} else {
		msg = fmt.Sprint(first)
	}

	// Log message
	logger.LogAttrs(context.Background(), level, msg, slog.String("component", component))
}

// Info logs an info message
func Info(first interface{}, args ...interface{}) {
	logWithOptionalComponent(slog.LevelInfo, first, args...)
}

// Warn logs a warning message
func Warn(first interface{}, args ...interface{}) {
	logWithOptionalComponent(slog.LevelWarn, first, args...)
}

// Error logs an error message
func Error(first interface{}, args ...interface{}) {
	logWithOptionalComponent(slog.LevelError, first, args...)
}

// Debug logs a debug message
func Debug(first interface{}, args ...interface{}) {
	logWithOptionalComponent(slog.LevelDebug, first, args...)
}

// Fatal logs a fatal message
func Fatal(first interface{}, args ...interface{}) {
	logWithOptionalComponent(LevelFatal, first, args...)
	ExitFunc(1)
}

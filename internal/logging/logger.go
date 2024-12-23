package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

// logger is the global logger
var logger *slog.Logger

// ExitFunc represents the exit function (used for testing)
var ExitFunc = os.Exit

// ComponentType represents the type of component
type ComponentType string

// CustomTextHandler represents a custom text handler
type CustomTextHandler struct {
	slog.Handler
	out        io.Writer
	level      slog.Level
	levelNames map[slog.Level]string
}

const (
	APP   ComponentType = "[APP]"
	BLE   ComponentType = "[BLE]"
	SPEED ComponentType = "[SPD]"
	VIDEO ComponentType = "[VID]"
)

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

// Create a new slog level for the Fatal logging level
const LevelFatal slog.Level = slog.Level(12)

// ExitHandler is a function type for handling fatal exits
type ExitHandler func()

var (
	exitHandler ExitHandler
	exitOnce    sync.Once
)

// SetExitHandler sets the handler for fatal exits
func SetExitHandler(handler ExitHandler) {
	exitHandler = handler
}

// Initialize sets up the logger
func Initialize(logLevel string) *slog.Logger {
	level := parseLogLevel(logLevel)
	logger = slog.New(NewCustomTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	return logger
}

// Info logs an info message
func Info(first interface{}, args ...interface{}) {
	logWithOptionalComponent(context.Background(), slog.LevelInfo, first, args...)
}

// Warn logs a warning message
func Warn(first interface{}, args ...interface{}) {
	logWithOptionalComponent(context.Background(), slog.LevelWarn, first, args...)
}

// Error logs an error message
func Error(first interface{}, args ...interface{}) {
	logWithOptionalComponent(context.Background(), slog.LevelError, first, args...)
}

// Debug logs a debug message
func Debug(first interface{}, args ...interface{}) {
	logWithOptionalComponent(context.Background(), slog.LevelDebug, first, args...)
}

// Fatal logs a fatal message and triggers exit handler
func Fatal(first interface{}, args ...interface{}) {
	logWithOptionalComponent(context.Background(), LevelFatal, first, args...)

	// Trigger exit handler
	exitOnce.Do(func() {

		if exitHandler != nil {
			exitHandler()
		}

		ExitFunc(0)
	})
}

// NewCustomTextHandler creates a new custom text handler
func NewCustomTextHandler(w io.Writer, opts *slog.HandlerOptions) *CustomTextHandler {
	// Set default values if not provided
	if w == nil {
		w = os.Stdout
	}

	if opts == nil {
		opts = &slog.HandlerOptions{Level: slog.LevelInfo}
	}

	levelNames := map[slog.Level]string{
		slog.LevelDebug: "DBG",
		slog.LevelInfo:  "INF",
		slog.LevelWarn:  "WRN",
		slog.LevelError: "ERR",
		LevelFatal:      "FTL",
	}

	// Create the custom text handler
	textHandler := slog.NewTextHandler(w, opts)
	return &CustomTextHandler{
		Handler:    textHandler,
		out:        w,
		level:      opts.Level.(slog.Level),
		levelNames: levelNames,
	}
}

// Handle handles the log record
func (h *CustomTextHandler) Handle(ctx context.Context, r slog.Record) error {
	// Check if context is done
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Create custom logger output
	timestamp := r.Time.Format("2006/01/02 15:04:05")
	level := strings.TrimSpace("[" + h.levelNames[r.Level] + "]")

	msg := r.Message

	// Write output format to writer
	fmt.Fprintf(h.out, "%s %s%s %s%s%s%s\n",
		timestamp,
		h.getColorForLevel(r.Level),
		level,
		Reset,
		h.getComponentFromAttrs(r),
		msg,
		Reset)
	return nil
}

// Enabled checks if the handler is enabled
func (h *CustomTextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Handler.Enabled(ctx, level)
}

// WithAttrs adds attributes to the handler
func (h *CustomTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CustomTextHandler{
		Handler:    h.Handler.WithAttrs(attrs),
		out:        h.out,
		level:      h.level,
		levelNames: h.levelNames,
	}
}

// WithGroup adds a group to the handler
func (h *CustomTextHandler) WithGroup(name string) slog.Handler {
	return &CustomTextHandler{
		Handler:    h.Handler.WithGroup(name),
		out:        h.out,
		level:      h.level,
		levelNames: h.levelNames,
	}
}

// getColorForLevel returns the color for the given log level
func (h *CustomTextHandler) getColorForLevel(level slog.Level) string {

	// Make color assignments based on log level
	switch level {
	case slog.LevelDebug:
		return Blue
	case slog.LevelInfo:
		return Green
	case slog.LevelWarn:
		return Yellow
	case slog.LevelError:
		return Magenta
	case LevelFatal:
		return Red
	default:
		return White
	}

}

// getComponentFromAttrs extracts and formats the component from record attributes
func (h *CustomTextHandler) getComponentFromAttrs(r slog.Record) string {
	var component string

	// Extract optional component from attributes
	r.Attrs(func(a slog.Attr) bool {

		if a.Key == "component" {
			component = a.Value.String()

			if component != "" {
				component = component + " "
			}

			return false
		}

		return true
	})

	return component
}

// parseLogLevel converts a string log level to slog.Level
func parseLogLevel(level string) slog.Level {

	// Convert log level to slog.Level
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo // default to Info level
	}

}

// logWithOptionalComponent logs a message with an optional component
func logWithOptionalComponent(ctx context.Context, level slog.Level, first interface{}, args ...interface{}) {
	// Check if context is nil
	if ctx == nil {
		ctx = context.Background()
	}

	var msg string
	var component string

	// Check if the first argument is an optional ComponentType
	switch v := first.(type) {
	case ComponentType:
		component = string(v)
		if len(args) > 0 {
			msg = fmt.Sprint(args[0])
		}
	default:
		msg = fmt.Sprint(first)
	}

	msg = strings.TrimSpace(msg)
	logger.LogAttrs(ctx, level, msg, slog.String("component", component))
}

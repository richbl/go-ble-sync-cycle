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

// ComponentType represents the type of component for logging identification
type ComponentType string

// Supported component types
const (
	APP   ComponentType = "[APP]"
	BLE   ComponentType = "[BLE]"
	SPEED ComponentType = "[SPD]"
	VIDEO ComponentType = "[VID]"
)

// ANSI color codes for terminal output
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

// LevelFatal defines a new slog level for fatal errors
const LevelFatal slog.Level = slog.Level(12)

// Global variables
var (
	logger      *slog.Logger
	ExitFunc    = os.Exit // ExitFunc represents the exit function (used for testing)
	exitHandler ExitHandler
	exitOnce    sync.Once
)

// Type definitions
type (
	// ExitHandler is a function type for handling fatal exits
	ExitHandler func()

	// CustomTextHandler represents a custom text handler for log formatting
	CustomTextHandler struct {
		slog.Handler
		out        io.Writer
		level      slog.Level
		levelNames map[slog.Level]string
	}
)

// Initialize sets up the logger with the specified log level
func Initialize(logLevel string) *slog.Logger {

	level := parseLogLevel(logLevel)
	logger = slog.New(NewCustomTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	return logger
}

// SetExitHandler sets the handler for fatal exits
func SetExitHandler(handler ExitHandler) {
	exitHandler = handler
}

// Logging functions
func Debug(first any, args ...any) {
	logWithOptionalComponent(context.Background(), slog.LevelDebug, first, args...)
}

func Info(first any, args ...any) {
	logWithOptionalComponent(context.Background(), slog.LevelInfo, first, args...)
}

func Warn(first any, args ...any) {
	logWithOptionalComponent(context.Background(), slog.LevelWarn, first, args...)
}

func Error(first any, args ...any) {
	logWithOptionalComponent(context.Background(), slog.LevelError, first, args...)
}

func Fatal(first any, args ...any) {

	logWithOptionalComponent(context.Background(), LevelFatal, first, args...)

	// Since we are exiting, we need to ensure that the exit handler is called
	exitOnce.Do(func() {

		if exitHandler != nil {
			exitHandler()
		}

		ExitFunc(0)
	})
}

// NewCustomTextHandler creates a new custom text handler with the specified options
func NewCustomTextHandler(w io.Writer, opts *slog.HandlerOptions) *CustomTextHandler {

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

	var level slog.Level

	level, ok := opts.Level.(slog.Level)
	if !ok {
		level = slog.LevelDebug
	}

	return &CustomTextHandler{
		Handler:    slog.NewTextHandler(w, opts),
		out:        w,
		level:      level,
		levelNames: levelNames,
	}
}

// Handle handles the log record and writes it to the output stream
func (h *CustomTextHandler) Handle(ctx context.Context, r slog.Record) error {

	if err := ctx.Err(); err != nil {
		return err
	}

	timestamp := r.Time.Format("2006/01/02 15:04:05")
	level := strings.TrimSpace("[" + h.levelNames[r.Level] + "]")

	// Write the formatted log record to the output
	if _, err := fmt.Fprintf(h.out, "%s %s%s %s%s%s%s\n",
		timestamp,
		h.getColorForLevel(r.Level),
		level,
		Reset,
		h.getComponentFromAttrs(r),
		r.Message,
		Reset); err != nil {
		return err
	}

	return nil
}

// Enabled returns true if the specified log level is enabled
func (h *CustomTextHandler) Enabled(ctx context.Context, level slog.Level) bool {

	return h.Handler.Enabled(ctx, level)
}

// WithAttrs returns a new handler with the specified attributes
func (h *CustomTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {

	return &CustomTextHandler{
		Handler:    h.Handler.WithAttrs(attrs),
		out:        h.out,
		level:      h.level,
		levelNames: h.levelNames,
	}
}

// WithGroup returns a new handler with the specified group name
func (h *CustomTextHandler) WithGroup(name string) slog.Handler {

	return &CustomTextHandler{
		Handler:    h.Handler.WithGroup(name),
		out:        h.out,
		level:      h.level,
		levelNames: h.levelNames,
	}
}

// getColorForLevel returns the ANSI color code for the specified log level
func (h *CustomTextHandler) getColorForLevel(level slog.Level) string {

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

// getComponentFromAttrs returns the component name from the log record attributes
func (h *CustomTextHandler) getComponentFromAttrs(r slog.Record) string {

	var component string

	r.Attrs(func(a slog.Attr) bool {

		if a.Key == "component" {
			component = a.Value.String()

			if component != "" {
				component += " "
			}

			return false
		}

		return true
	})

	return component
}

// parseLogLevel converts the specified log level string to a slog.Level
func parseLogLevel(level string) slog.Level {

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
		return slog.LevelInfo
	}

}

// buildMessage constructs the log message from the given argument
func buildMessage(first any, args ...any) string {

	var parts []string

	// Handle the first argument if it's not nil
	if first != nil {
		parts = append(parts, fmt.Sprintf("%v", first))
	}

	// Build the rest of the arguments
	for _, arg := range args {
		parts = append(parts, fmt.Sprintf("%v", arg))
	}

	return strings.Join(parts, " ")
}

// logWithOptionalComponent logs a message with an optional component name
func logWithOptionalComponent(ctx context.Context, level slog.Level, first any, args ...any) {

	var component string
	var msg string

	// Check if the first argument is a ComponentType and extract it
	if c, ok := first.(ComponentType); ok {
		component = string(c)
		msg = buildMessage(nil, args...)
	} else {
		msg = buildMessage(first, args...)
	}

	// Log the message with attributes
	logger.LogAttrs(ctx, level, msg, slog.String("component", component))
}

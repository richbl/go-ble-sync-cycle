package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

// ComponentType represents the types of component for logger identification
type ComponentType string

// Component types
const (
	APP   ComponentType = "[APP]"
	BLE   ComponentType = "[BLE]"
	SPEED ComponentType = "[SPD]"
	VIDEO ComponentType = "[VID]"
	GUI   ComponentType = "[GUI]"
)

// ANSI color codes used to distinguish log levels
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

type (
	ExitHandler func()

	CustomTextHandler struct {
		slog.Handler
		out        io.Writer
		levelNames map[slog.Level]string
	}

	syncMultiWriter struct {
		mu      sync.Mutex
		writers []io.Writer
	}
)

var (
	logger       *slog.Logger
	ExitFunc     = os.Exit
	exitHandler  ExitHandler
	exitOnce     sync.Once
	logOutput    *syncMultiWriter
	logLevelVar  = new(slog.LevelVar)
	outputFormat = "%s%s%s "
)

// Initialize sets up the logger
func Initialize(logLevel string) *slog.Logger {

	// Set the initial logging level value
	logLevelVar.Set(parseLogLevel(logLevel))

	// Set up the logger output
	logOutput = &syncMultiWriter{
		writers: []io.Writer{os.Stdout},
	}

	handler := NewCustomTextHandler(logOutput, &slog.HandlerOptions{Level: logLevelVar})
	logger = slog.New(handler)
	slog.SetDefault(logger)

	return logger
}

// AddWriter allows external components (like the GUI) to attach a log listener
func AddWriter(w io.Writer) {

	if logOutput != nil {
		logOutput.Add(w)
	}

}

// SetLogLevel dynamically updates the logging level of the running application
func SetLogLevel(levelStr string) {

	newLevel := parseLogLevel(levelStr)
	logLevelVar.Set(newLevel)

	Debug(APP, fmt.Sprintf("logging level changed to %s", newLevel.String()))

}

// LogLevel returns the current logging level
func LogLevel() string {
	return logLevelVar.Level().String()
}

// ClearCLILine clears the CLI
func ClearCLILine() {
	fmt.Print("\r\033[K")
}

// SetExitHandler sets the exit handler for fatal log events
func SetExitHandler(handler ExitHandler) {
	exitHandler = handler
}

// Logger functions for each log level

func Debug(first any, args ...any) {
	logStrict(context.Background(), slog.LevelDebug, first, args...)
}

func Info(first any, args ...any) {
	logStrict(context.Background(), slog.LevelInfo, first, args...)
}

func Warn(first any, args ...any) {
	logStrict(context.Background(), slog.LevelWarn, first, args...)
}

func Error(first any, args ...any) {
	logStrict(context.Background(), slog.LevelError, first, args...)
}

// Fatal logs a fatal error and exits
func Fatal(first any, args ...any) {

	logStrict(context.Background(), LevelFatal, first, args...)

	exitOnce.Do(func() {

		// Check to see if something needs to be done before punching out
		if exitHandler != nil {
			exitHandler()
		}

	})

	ExitFunc(1)

}

// logStrict formats and outputs a log message
func logStrict(ctx context.Context, level slog.Level, first any, args ...any) {

	var msg string
	var attrs []any

	// Check if the first argument is a component
	if c, ok := first.(ComponentType); ok {
		attrs = append(attrs, slog.String("_component", string(c)))

		// Check for remaining arguments
		if len(args) > 0 {
			msg = fmt.Sprint(args[0])
			attrs = append(attrs, args[1:]...)
		}

		// First argument is not a component, so use it as the message element
	} else {
		msg = fmt.Sprint(first)
		attrs = append(attrs, args...)
	}

	logger.Log(ctx, level, msg, attrs...)

}

// parseLogLevel converts a string to a slog.Level
func parseLogLevel(level string) slog.Level {

	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}

}

// syncMultiWriter is a custom writer that allows multiple writers to be added
func (m *syncMultiWriter) Add(w io.Writer) {

	m.mu.Lock()
	defer m.mu.Unlock()

	m.writers = append(m.writers, w)

}

// Write writes to all attached writers
func (m *syncMultiWriter) Write(p []byte) (n int, err error) {

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, w := range m.writers {
		n, err = w.Write(p)
		if err != nil {
			return n, err
		}
	}

	return len(p), nil
}

// NewCustomTextHandler creates a new CustomTextHandler
func NewCustomTextHandler(out io.Writer, opts *slog.HandlerOptions) *CustomTextHandler {

	return &CustomTextHandler{
		Handler: slog.NewTextHandler(out, opts),
		out:     out,
		levelNames: map[slog.Level]string{
			slog.LevelDebug: "[DBG]",
			slog.LevelInfo:  "[INF]",
			slog.LevelWarn:  "[WRN]",
			slog.LevelError: "[ERR]",
			LevelFatal:      "[FTL]",
		},
	}
}

// Handle formats and outputs a log message
func (h *CustomTextHandler) Handle(ctx context.Context, r slog.Record) error {

	var buf bytes.Buffer

	// Set the timestamp
	fmt.Fprintf(&buf, outputFormat, White, r.Time.Format("15:04:05"), Reset)

	// Set the log level
	h.appendLevel(&buf, r.Level)

	// Get the component and attributes
	component, otherAttrs := h.extractComponentAndAttrs(r)

	if component != "" {
		fmt.Fprintf(&buf, outputFormat, Blue, component, Reset)
	}

	// Set the message in the buffer
	fmt.Fprintf(&buf, "%s", r.Message)

	// Append the attributes to the buffer
	h.appendAttrs(&buf, otherAttrs)

	// Finally write the buffer
	buf.WriteString("\n")
	_, err := h.out.Write(buf.Bytes())

	return err
}

// appendLevel appends the log level formatting to the buffer
func (h *CustomTextHandler) appendLevel(buf *bytes.Buffer, level slog.Level) {

	levelName := h.levelNames[level]
	if levelName == "" {
		levelName = fmt.Sprintf("[%s]", level.String())
	}

	// Set the level color
	var levelColor string

	switch level {
	case slog.LevelDebug:
		levelColor = Cyan
	case slog.LevelInfo:
		levelColor = Green
	case slog.LevelWarn:
		levelColor = Yellow
	case slog.LevelError:
		levelColor = Red
	case LevelFatal:
		levelColor = Magenta
	default:
		levelColor = White
	}

	fmt.Fprintf(buf, outputFormat, levelColor, levelName, Reset)

}

// extractComponentAndAttrs returns the component and attributes from the record
func (h *CustomTextHandler) extractComponentAndAttrs(r slog.Record) (string, []slog.Attr) {

	var component string
	var attrs []slog.Attr

	// Iterate over the attributes
	r.Attrs(func(a slog.Attr) bool {

		if a.Key == "_component" {
			component = a.Value.String()
			return true // Continue getting other attrs
		}

		attrs = append(attrs, a)

		return true
	})

	return component, attrs
}

// appendAttrs appends the attributes to the buffer
func (h *CustomTextHandler) appendAttrs(buf *bytes.Buffer, attrs []slog.Attr) {

	if len(attrs) == 0 {
		return
	}

	buf.WriteString(" ") // Add a space before the attributes

	// Iterate over the attributes
	for i, a := range attrs {
		fmt.Fprintf(buf, "%s%s=%v%s", White, a.Key, a.Value, Reset)

		if i < len(attrs)-1 {
			buf.WriteString(" ")
		}

	}

}

// UseGUIWriterOnly replaces all writers with only the GUI writer (used in GUI mode).
func UseGUIWriterOnly(w io.Writer) {
	logOutput.mu.Lock()
	defer logOutput.mu.Unlock()
	logOutput.writers = []io.Writer{w}
}

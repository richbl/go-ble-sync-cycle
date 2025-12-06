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

// ComponentType represents the type of component for logger identification
type ComponentType string

// Supported component types
const (
	APP   ComponentType = "[APP]"
	BLE   ComponentType = "[BLE]"
	SPEED ComponentType = "[SPD]"
	VIDEO ComponentType = "[VID]"
	GUI   ComponentType = "[GUI]"
)

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

// LevelFatal defines a new slog level for fatal errors
const LevelFatal slog.Level = slog.Level(12)

var (
	logger      *slog.Logger
	exitFunc    = os.Exit
	exitHandler ExitHandler
	exitOnce    sync.Once
	logOutput   *syncMultiWriter
)

type (
	ExitHandler func()

	CustomTextHandler struct {
		slog.Handler
		out        io.Writer
		level      slog.Level
		levelNames map[slog.Level]string
	}

	syncMultiWriter struct {
		mu      sync.Mutex
		writers []io.Writer
	}
)

// Initialize the logger
func Initialize(logLevel string) *slog.Logger {

	level := parseLogLevel(logLevel)

	logOutput = &syncMultiWriter{
		writers: []io.Writer{os.Stdout},
	}

	handler := NewCustomTextHandler(logOutput, &slog.HandlerOptions{Level: level})
	logger = slog.New(handler)
	slog.SetDefault(logger)

	return logger
}

// SetExitHandler sets the exit handler
func SetExitHandler(handler ExitHandler) {
	exitHandler = handler
}

// Wrapper functions
// Pattern: ([Component], Message, [Key, Value...])

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

func Fatal(first any, args ...any) {

	logStrict(context.Background(), LevelFatal, first, args...)

	// Call exit handler to allow for clean shutdown logic
	exitOnce.Do(func() {

		if exitHandler != nil {
			exitHandler()
		}

	})
	exitFunc(1)

}

// logStrict is a strict logger
func logStrict(ctx context.Context, level slog.Level, first any, args ...any) {

	var msg string
	var attrs []any

	// Extract component and attributes
	if c, ok := first.(ComponentType); ok {

		// First arg is the component type
		attrs = append(attrs, slog.String("_component", string(c)))

		if len(args) > 0 {
			msg = fmt.Sprint(args[0])
			attrs = append(attrs, args[1:]...)
		}

	} else {
		// First arg is the message
		msg = fmt.Sprint(first)
		attrs = append(attrs, args...)
	}

	// Now generate the log entry
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

// syncMultiWriter is a thread-safe multi writer for the logger
func (m *syncMultiWriter) Add(w io.Writer) {

	m.mu.Lock()
	defer m.mu.Unlock()
	m.writers = append(m.writers, w)

}

// Write implements the io.Writer interface
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

// NewCustomTextHandler creates a new custom text handler
func NewCustomTextHandler(out io.Writer, opts *slog.HandlerOptions) *CustomTextHandler {

	return &CustomTextHandler{
		Handler: slog.NewTextHandler(out, opts),
		out:     out,
		level:   opts.Level.Level(),
		levelNames: map[slog.Level]string{
			slog.LevelDebug: "[DBG]",
			slog.LevelInfo:  "[INF]",
			slog.LevelWarn:  "[WRN]",
			slog.LevelError: "[ERR]",
			LevelFatal:      "[FTL]",
		},
	}
}

// Enabled implements the slog.Handler interface
func (h *CustomTextHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle implements the slog.Handler interface
func (h *CustomTextHandler) Handle(ctx context.Context, r slog.Record) error {

	var buf bytes.Buffer

	// Set timestamp
	fmt.Fprintf(&buf, "%s%s%s ", White, r.Time.Format("15:04:05"), Reset)

	// Set level
	h.appendLevel(&buf, r.Level)

	// Extract component and attributes
	component, otherAttrs := h.extractComponentAndAttrs(r)

	// Set component
	if component != "" {
		fmt.Fprintf(&buf, "%s%s%s ", Blue, component, Reset)
	}

	// Create the message
	fmt.Fprintf(&buf, "%s", r.Message)

	// Append attributes
	h.appendAttrs(&buf, otherAttrs)

	// Add newline
	buf.WriteString("\n")
	_, err := h.out.Write(buf.Bytes())

	return err
}

// appendLevel appends the level to the buffer
func (h *CustomTextHandler) appendLevel(buf *bytes.Buffer, level slog.Level) {

	levelName := h.levelNames[level]
	if levelName == "" {
		levelName = fmt.Sprintf("[%s]", level.String())
	}

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

	fmt.Fprintf(buf, "%s%s%s ", levelColor, levelName, Reset)

}

// extractComponentAndAttrs extracts the component and attributes from the record
func (h *CustomTextHandler) extractComponentAndAttrs(r slog.Record) (string, []slog.Attr) {

	var component string
	var attrs []slog.Attr

	r.Attrs(func(a slog.Attr) bool {

		if a.Key == "_component" {
			component = a.Value.String()

			return true
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

	buf.WriteString(" ")

	for i, a := range attrs {
		fmt.Fprintf(buf, "%s%s=%v%s", White, a.Key, a.Value, Reset)

		if i < len(attrs)-1 {
			buf.WriteString(" ")
		}

	}

}

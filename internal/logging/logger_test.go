package logger

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

// testWriter is a custom writer that captures output and allows inspection
type testWriter struct {
	buf bytes.Buffer
}

// Write implements io.Writer
func (w *testWriter) Write(p []byte) (n int, err error) {
	return w.buf.Write(p)
}

// String implements fmt.Stringer
func (w *testWriter) String() string {
	return w.buf.String()
}

// Reset resets the buffer
func (w *testWriter) Reset() {
	w.buf.Reset()
}

// setupTestLogger creates a new logger with a test writer
func setupTestLogger(t *testing.T, level string) (*testWriter, *slog.Logger) {
	t.Helper()
	writer := &testWriter{}
	originalLogger := logger // Save current logger
	logger = Initialize(level)

	// Create new handler with the test writer
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	logger = slog.New(NewCustomTextHandler(writer, &slog.HandlerOptions{
		Level: logLevel,
	}))

	return writer, originalLogger
}

// cleanupTestLogger restores the original logger
func cleanupTestLogger(t *testing.T, originalLogger *slog.Logger) {
	t.Helper()
	logger = originalLogger
}

func TestInitialize(t *testing.T) {

	tests := []struct {
		name          string
		level         string
		expectedLevel slog.Level
	}{
		{"debug level", "debug", slog.LevelDebug},
		{"info level", "info", slog.LevelInfo},
		{"warn level", "warn", slog.LevelWarn},
		{"error level", "error", slog.LevelError},
		{"default level", "invalid", slog.LevelInfo},
		{"empty level", "", slog.LevelInfo},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			logger := Initialize(tt.level)

			if logger == nil {
				t.Fatal("Expected non-nil logger")
			}

			// Test if logger is properly initialized with correct level
			ctx := context.Background()
			handler := logger.Handler()

			if !handler.Enabled(ctx, tt.expectedLevel) {
				t.Errorf("Expected level %v to be enabled", tt.expectedLevel)
			}

		})
	}
}

func TestLoggingLevels(t *testing.T) {

	tests := []struct {
		name          string
		logFunc       func(string, ...interface{})
		message       string
		expectedLevel string
		expectedColor string
	}{
		{"debug message", Debug, "debug test", "DEBUG", Blue},
		{"info message", Info, "info test", "INFO", Green},
		{"warn message", Warn, "warn test", "WARN", Yellow},
		{"error message", Error, "error test", "ERROR", Red},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			// Set up test logger
			writer, origLogger := setupTestLogger(t, "debug")
			defer cleanupTestLogger(t, origLogger)

			tt.logFunc(tt.message)
			output := writer.String()

			// Check if output contains expected components
			if !strings.Contains(output, tt.expectedLevel) {
				t.Errorf("Expected output to contain level %q, got: %q", tt.expectedLevel, output)
			}

			if !strings.Contains(output, tt.message) {
				t.Errorf("Expected output to contain message %q, got: %q", tt.message, output)
			}

			if !strings.Contains(output, tt.expectedColor) {
				t.Errorf("Expected output to contain color %q, got: %q", tt.expectedColor, output)
			}

		})
	}
}

func TestCustomTextHandler(t *testing.T) {

	// Create a custom text handler
	writer := &testWriter{}
	handler := NewCustomTextHandler(writer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	// Test basic handler functionality
	ctx := context.Background()
	record := slog.Record{
		Time:    time.Now(),
		Level:   slog.LevelInfo,
		Message: "test message",
	}

	err := handler.Handle(ctx, record)

	if err != nil {
		t.Errorf("Unexpected error handling record: %v", err)
	}

	output := writer.String()

	if !strings.Contains(output, "INFO") {
		t.Errorf("Expected output to contain 'INFO', got: %q", output)
	}

	if !strings.Contains(output, "test message") {
		t.Errorf("Expected output to contain 'test message', got: %q", output)
	}

}

func TestWithAttrsAndGroups(t *testing.T) {

	// Create a custom text handler
	writer := &testWriter{}
	handler := NewCustomTextHandler(writer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	// Test WithAttrs
	attrs := []slog.Attr{slog.String("key", "value")}
	handlerWithAttrs := handler.WithAttrs(attrs)

	if handlerWithAttrs == nil {
		t.Error("Expected non-nil handler with attributes")
	}

	// Test WithGroup
	handlerWithGroup := handler.WithGroup("group")

	if handlerWithGroup == nil {
		t.Error("Expected non-nil handler with group")
	}

}

func TestFatal(t *testing.T) {

	// Create a temporary executable that calls Fatal()
	if os.Getenv("TEST_FATAL") == "1" {
		writer := &testWriter{}

		logger = slog.New(NewCustomTextHandler(writer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		Fatal("fatal error")
		return
	}

	// Run the test in a subprocess
	cmd := os.Args[0]
	env := []string{"TEST_FATAL=1"}

	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Execute the test process
	p, err := os.StartProcess(cmd, []string{cmd, "-test.run=TestFatal"}, &os.ProcAttr{
		Env: append(os.Environ(), env...),
	})

	if err != nil {
		t.Fatal(err)
	}

	// Wait for the process to finish
	state, err := p.Wait()

	if err != nil {
		t.Fatal(err)
	}

	// Check if the process exited with status 1
	if code := state.ExitCode(); code != 1 {
		t.Errorf("Expected exit code 1, got %d", code)
	}

}

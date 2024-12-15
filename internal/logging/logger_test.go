package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

const (
	testMessage    = "message"
	defaultOptions = slog.LevelDebug
)

// testCase represents a generic test case for logging tests
type testCase struct {
	name     string
	level    interface{} // can be string or slog.Level
	want     interface{} // can be slog.Level or string
	setLevel slog.Level  // used only for Enabled tests
}

// setupTestLogger creates a new logger with a buffer for testing
func setupTestLogger() (*bytes.Buffer, *slog.Logger) {

	buf := &bytes.Buffer{}
	handler := NewCustomTextHandler(buf, &slog.HandlerOptions{Level: defaultOptions})
	return buf, slog.New(handler)

}

func TestInitialize(t *testing.T) {

	// Define test cases
	tests := []testCase{
		{"debug", "debug", slog.LevelDebug, 0},
		{"info", "info", slog.LevelInfo, 0},
		{"warn", "warn", slog.LevelWarn, 0},
		{"error", "error", slog.LevelError, 0},
		{"default", "unknown", slog.LevelInfo, 0},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Initialize(tt.level.(string))

			if logger == nil {
				t.Fatal("logger is nil")
			}

			h, ok := logger.Handler().(*CustomTextHandler)
			if !ok {
				t.Fatal("logger handler is not of type *CustomTextHandler")
			}

			if h.level != tt.want.(slog.Level) {
				t.Errorf("got logger level %v, want %v", h.level, tt.want)
			}
		})
	}
}

// TestCustomTextHandler tests the custom text handler
func TestCustomTextHandler(t *testing.T) {

	// Define test cases
	tests := []testCase{
		{"debug", slog.LevelDebug, "\033[34mDEBUG\033[0m", 0},
		{"info", slog.LevelInfo, "\033[32mINFO\033[0m", 0},
		{"warn", slog.LevelWarn, "\033[33mWARN\033[0m", 0},
		{"error", slog.LevelError, "\033[31mERROR\033[0m", 0},
		{"fatal", LevelFatal, "\033[35mFATAL\033[0m", 0},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			h := NewCustomTextHandler(buf, &slog.HandlerOptions{Level: defaultOptions})
			r := slog.NewRecord(time.Now(), tt.level.(slog.Level), testMessage, 0)

			if err := h.Handle(context.Background(), r); err != nil {
				t.Fatalf("Handle() error = %v", err)
			}
			if !strings.Contains(buf.String(), tt.want.(string)) {
				t.Errorf("got %q, want %q", buf.String(), tt.want)
			}
		})
	}
}

// TestLogLevels tests all logging level functions
func TestLogLevels(t *testing.T) {

	// Define test cases
	tests := []struct {
		name      string
		logFunc   func(first interface{}, args ...interface{})
		wantLevel string
	}{
		{"Info", Info, "INFO"},
		{"Warn", Warn, "WARN"},
		{"Error", Error, "ERROR"},
		{"Debug", Debug, "DEBUG"},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf, testLogger := setupTestLogger()
			logger = testLogger

			tt.logFunc(testMessage)
			if !strings.Contains(buf.String(), tt.wantLevel) {
				t.Errorf("got %q, want to contain %q", buf.String(), tt.wantLevel)
			}
		})
	}
}

// TestFatal tests the Fatal function
func TestFatal(t *testing.T) {

	buf, testLogger := setupTestLogger()
	logger = testLogger

	// Save and restore exit function so we can check if it was called
	savedExitFunc := ExitFunc
	defer func() { ExitFunc = savedExitFunc }()
	exitCalled := false

	// Mock exit function
	ExitFunc = func(code int) {
		exitCalled = true
		if code != 1 {
			t.Errorf("Fatal called exit function with code %d, want 1", code)
		}
	}

	// Call Fatal function for testing
	Fatal(testMessage)
	if !exitCalled {
		t.Error("Fatal did not call exit function")
	}
	if buf.String() == "" {
		t.Error("Fatal did not log a message")
	}
}

// TestEnabled tests the Enabled function
func TestEnabled(t *testing.T) {

	// Define test cases
	tests := []testCase{
		{"debug enabled", slog.LevelDebug, true, slog.LevelDebug},
		{"info enabled", slog.LevelInfo, true, slog.LevelDebug},
		{"warn enabled", slog.LevelWarn, true, slog.LevelDebug},
		{"error enabled", slog.LevelError, true, slog.LevelDebug},
		{"fatal enabled", LevelFatal, true, slog.LevelDebug},
		{"debug disabled", slog.LevelDebug, false, slog.LevelInfo},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewCustomTextHandler(&bytes.Buffer{}, &slog.HandlerOptions{Level: tt.setLevel})

			if got := h.Enabled(context.Background(), tt.level.(slog.Level)); got != tt.want.(bool) {
				t.Errorf("Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

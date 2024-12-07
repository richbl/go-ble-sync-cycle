package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

// TestInitialize tests the Initialize function
func TestInitialize(t *testing.T) {

	// Define test cases
	tests := []struct {
		name  string
		level string
		want  slog.Level
	}{
		{"debug", "debug", slog.LevelDebug},
		{"info", "info", slog.LevelInfo},
		{"warn", "warn", slog.LevelWarn},
		{"error", "error", slog.LevelError},
		{"default", "unknown", slog.LevelInfo},
	}

	// Run tests for each log level
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			Initialize(tt.level)

			if logger == nil {
				t.Errorf("logger is nil")
			}

			h, ok := logger.Handler().(*CustomTextHandler)
			if !ok {
				t.Errorf("logger handler is not of type *CustomTextHandler")
			}

			if h.level != tt.want {
				t.Errorf("got logger level %v, want %v", h.level, tt.want)
			}

		})
	}

}

func TestCustomTextHandler(t *testing.T) {

	// Define test cases
	tests := []struct {
		name  string
		level slog.Level
		want  string
	}{
		{"debug", slog.LevelDebug, "\033[34mDEBUG\033[0m"},
		{"info", slog.LevelInfo, "\033[32mINFO\033[0m"},
		{"warn", slog.LevelWarn, "\033[33mWARN\033[0m"},
		{"error", slog.LevelError, "\033[31mERROR\033[0m"},
		{"fatal", LevelFatal, "\033[35mFATAL\033[0m"},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			var buf bytes.Buffer
			h := NewCustomTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
			r := slog.NewRecord(time.Now(), tt.level, "message", 0)

			if err := h.Handle(context.Background(), r); err != nil {
				t.Errorf("Handle() error = %v", err)
			}

			if !strings.Contains(buf.String(), tt.want) {
				t.Errorf("got %q, want %q", buf.String(), tt.want)
			}

		})
	}

}

// TestInfo tests the INFO log level function
func TestInfo(t *testing.T) {

	var buf bytes.Buffer
	Initialize("debug")
	logger = slog.New(NewCustomTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	Info("message")
	if !strings.Contains(buf.String(), "INFO") {
		t.Errorf("got %q, want %q", buf.String(), "INFO")
	}

}

// TestWarn tests the WARN log level function
func TestWarn(t *testing.T) {

	var buf bytes.Buffer
	Initialize("debug")
	logger = slog.New(NewCustomTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	Warn("message")
	if !strings.Contains(buf.String(), "WARN") {
		t.Errorf("got %q, want %q", buf.String(), "WARN")
	}

}

// TestError tests the ERROR log level function
func TestError(t *testing.T) {

	var buf bytes.Buffer
	Initialize("debug")
	logger = slog.New(NewCustomTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	Error("message")
	if !strings.Contains(buf.String(), "ERROR") {
		t.Errorf("got %q, want %q", buf.String(), "ERROR")
	}

}

// TestDebug tests the DEBUG log level function
func TestDebug(t *testing.T) {

	var buf bytes.Buffer
	Initialize("debug")
	logger = slog.New(NewCustomTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	Debug("message")
	if !strings.Contains(buf.String(), "DEBUG") {
		t.Errorf("got %q, want %q", buf.String(), "DEBUG")
	}

}

// TestFatal tests the FATAL log level function
func TestFatal(t *testing.T) {

	var buf bytes.Buffer
	Initialize("debug")
	logger = slog.New(NewCustomTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Save and restore exit function
	savedExitFunc := ExitFunc
	defer func() {
		ExitFunc = savedExitFunc
	}()

	called := false
	ExitFunc = func(code int) {
		called = true

		if code != 1 {
			t.Errorf("Fatal called exit function with code %d, want 1", code)
		}
	}

	Fatal("message")
	if !called {
		t.Errorf("Fatal did not call exit function")
	}

	if buf.String() == "" {
		t.Errorf("Fatal did not log a message")
	}

}

// TestEnabled tests the Enabled state of the log levels
func TestEnabled(t *testing.T) {

	// Define test cases
	tests := []struct {
		name  string
		level slog.Level
		want  bool
	}{
		{"debug", slog.LevelDebug, true},
		{"info", slog.LevelInfo, true},
		{"warn", slog.LevelWarn, true},
		{"error", slog.LevelError, true},
		{"fatal", LevelFatal, true},
		{"disabled", slog.LevelDebug, false},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			var buf bytes.Buffer
			h := NewCustomTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})

			if tt.name == "disabled" {
				h = NewCustomTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
			}

			if got := h.Enabled(context.Background(), tt.level); got != tt.want {
				t.Errorf("Enabled() = %v, want %v", got, tt.want)
			}

		})
	}

}

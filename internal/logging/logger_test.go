package logger

import (
	"bytes"
	"context"
	"log/slog"
	"regexp"
	"strings"
	"testing"
	"time"
)

// testData holds all test constants
type testData struct {
	message     string
	level       slog.Level
	colorCodes  map[slog.Level]string
	defaultOpts slog.Level
}

// Define test data
var td = testData{
	message:     "test message",
	level:       slog.LevelDebug,
	defaultOpts: slog.LevelDebug,
	colorCodes: map[slog.Level]string{
		slog.LevelDebug: "\033[34m[DEBUG]",
		slog.LevelInfo:  "\033[32m[INFO]",
		slog.LevelWarn:  "\033[33m[WARN]",
		slog.LevelError: "\033[31m[ERROR]",
		LevelFatal:      "\033[35mFATAL",
	},
}

// testCase represents a generic test case
type testCase struct {
	name     string
	level    slog.Level
	want     interface{}
	setLevel slog.Level
}

// setupTest creates a new test logger with buffer
func setupTest() (*bytes.Buffer, *slog.Logger) {
	buf := &bytes.Buffer{}
	handler := NewCustomTextHandler(buf, &slog.HandlerOptions{Level: td.level})
	return buf, slog.New(handler)
}

// validateLogOutput checks if log output matches expected format
func validateLogOutput(t *testing.T, output, expectedLevel string) {
	t.Helper()

	timestampRegex := `^\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}`
	if !regexp.MustCompile(timestampRegex).MatchString(output) {
		t.Errorf("invalid timestamp format in output: %q", output)
	}

	if !strings.Contains(output, expectedLevel) {
		t.Errorf("output %q missing expected level %q", output, expectedLevel)
	}

	if !strings.Contains(output, td.message) {
		t.Errorf("output %q missing message %q", output, td.message)
	}
}

func TestInitialize(t *testing.T) {
	// Define test cases
	tests := []struct {
		name      string
		level     string
		wantLevel slog.Level
	}{
		{"debug level", "debug", slog.LevelDebug},
		{"info level", "info", slog.LevelInfo},
		{"warn level", "warn", slog.LevelWarn},
		{"error level", "error", slog.LevelError},
		{"invalid level", "invalid", slog.LevelInfo}, // defaults to info
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Initialize(tt.level)

			if logger == nil {
				t.Fatal("logger is nil")
			}

			h, ok := logger.Handler().(*CustomTextHandler)
			if !ok {
				t.Fatal("invalid handler type")
			}

			if h.level != tt.wantLevel {
				t.Errorf("got level %v, want %v", h.level, tt.wantLevel)
			}
		})
	}
}

func TestCustomTextHandler(t *testing.T) {
	// Define test cases
	tests := []testCase{
		{"debug", slog.LevelDebug, "\033[34m[DEBUG] \033[0m", 0},
		{"info", slog.LevelInfo, "\033[32m[INFO] \033[0m", 0},
		{"warn", slog.LevelWarn, "\033[33m[WARN] \033[0m", 0},
		{"error", slog.LevelError, "\033[31m[ERROR] \033[0m", 0},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			h := NewCustomTextHandler(buf, &slog.HandlerOptions{Level: td.defaultOpts})
			r := slog.NewRecord(time.Now(), tt.level, td.message, 0)

			if err := h.Handle(context.Background(), r); err != nil {
				t.Fatalf("Handle() error = %v", err)
			}

			output := buf.String()
			expectedLevel := tt.want.(string)

			timestampRegex := `^\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2} `
			if !regexp.MustCompile(timestampRegex).MatchString(output) {
				t.Errorf("output does not start with a valid timestamp")
			}

			if !strings.Contains(output, expectedLevel) {
				t.Errorf("output %q does not contain expected level %q", output, expectedLevel)
			}

			if !strings.Contains(output, td.message) {
				t.Errorf("output %q does not contain message %q", output, td.message)
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	// Define test cases
	tests := []struct {
		name    string
		logFunc func(interface{}, ...interface{})
		level   string
	}{
		{"Debug", Debug, "DEBUG"},
		{"Info", Info, "INFO"},
		{"Warn", Warn, "WARN"},
		{"Error", Error, "ERROR"},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf, testLogger := setupTest()
			logger = testLogger

			tt.logFunc(td.message)
			validateLogOutput(t, buf.String(), tt.level)
		})
	}
}

func TestFatal(t *testing.T) {
	buf, testLogger := setupTest()
	logger = testLogger

	origExit := ExitFunc
	exitCode := 0
	ExitFunc = func(code int) { exitCode = code }
	defer func() { ExitFunc = origExit }()

	Fatal(td.message)

	if exitCode != 1 {
		t.Errorf("got exit code %d, want 1", exitCode)
	}

	validateLogOutput(t, buf.String(), "FATAL")
}

func TestEnabled(t *testing.T) {
	// Define test cases
	tests := []testCase{
		{"debug enabled", slog.LevelDebug, true, slog.LevelDebug},
		{"info disabled", slog.LevelInfo, false, slog.LevelError},
		{"error enabled", slog.LevelError, true, slog.LevelDebug},
		{"fatal enabled", LevelFatal, true, slog.LevelDebug},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewCustomTextHandler(&bytes.Buffer{}, &slog.HandlerOptions{Level: tt.setLevel})
			if got := h.Enabled(context.Background(), tt.level); got != tt.want.(bool) {
				t.Errorf("Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

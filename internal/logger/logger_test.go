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
	defaultOpts slog.Level
}

// testCase represents a generic test case
type testCase struct {
	name     string
	level    slog.Level
	want     bool
	setLevel slog.Level
}

// Define test data
var td = testData{
	message:     testMessage,
	level:       slog.LevelDebug,
	defaultOpts: slog.LevelDebug,
}

var testMessage = "test message"

// setupTest creates a new test logger with buffer
func setupTest() (*bytes.Buffer, *slog.Logger) {

	buf := &bytes.Buffer{}
	handler := NewCustomTextHandler(buf, &slog.HandlerOptions{Level: td.level})

	return buf, slog.New(handler)
}

// validateLogOutput checks if log output matches expected format (ANSI tolerant)
func validateLogOutput(t *testing.T, output, expectedLevel string) {

	t.Helper()

	// Matches optional ANSI-formatted start color, time, optional reset color
	timestampRegex := `^(\x1b\[[0-9;]*m)?\d{2}:\d{2}:\d{2}(\x1b\[[0-9;]*m)?`

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

// TestInitialize tests the initialization of the logger
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
		{"invalid level", "invalid", slog.LevelInfo},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Initialize(tt.level)

			if logger == nil {
				t.Fatal("logger is nil")
			}

			// Verify the logging level was updated correctly
			if logLevelVar.Level() != tt.wantLevel {
				t.Errorf("Initialize(%s) set level to %v, want %v", tt.level, logLevelVar.Level(), tt.wantLevel)
			}

		})
	}

}

// TestCustomTextHandler tests the custom text handler formatting and colors
func TestCustomTextHandler(t *testing.T) {

	// Define test cases
	tests := []struct {
		name     string
		level    slog.Level
		expected string
	}{
		// Expect ANSI color codes in the output
		{"debug", slog.LevelDebug, Cyan + "[DBG]"},
		{"info", slog.LevelInfo, Green + "[INF]"},
		{"warn", slog.LevelWarn, Yellow + "[WRN]"},
		{"error", slog.LevelError, Red + "[ERR]"},
	}

	// Run tests
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			buf := &bytes.Buffer{}
			h := NewCustomTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})

			// Create a log record
			r := slog.NewRecord(time.Now(), tt.level, testMessage, 0)

			// Handle the log record
			if err := h.Handle(context.Background(), r); err != nil {
				t.Fatalf("Handle() error = %v", err)
			}

			assertOutput(t, buf.String(), tt.expected, testMessage)
		})
	}

}

// assertOutput checks if the output contains the expected timestamp, level, and message
func assertOutput(t *testing.T, output, expectedLevel, expectedMessage string) {

	t.Helper()

	// Check timestamp (ANSI tolerant)
	// Logic: optional ANSI -> Digits -> Optional ANSI -> Space
	timestampRegex := `^(\x1b\[[0-9;]*m)?\d{2}:\d{2}:\d{2}(\x1b\[[0-9;]*m)? `

	if !regexp.MustCompile(timestampRegex).MatchString(output) {
		t.Errorf("output does not start with a valid timestamp format. Got: %q", output)
	}

	if !strings.Contains(output, expectedLevel) {
		t.Errorf("output %q does not contain expected level %q", output, expectedLevel)
	}

	if !strings.Contains(output, expectedMessage) {
		t.Errorf("output %q does not contain message %q", output, expectedMessage)
	}

}

// TestLogLevels tests the log level wrapper functions
func TestLogLevels(t *testing.T) {

	// Define test cases
	tests := []struct {
		name    string
		logFunc func(any, ...any)
		level   string
	}{
		{"Debug", Debug, "DBG"},
		{"Info", Info, "INF"},
		{"Warn", Warn, "WRN"},
		{"Error", Error, "ERR"},
	}

	// Run tests
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			buf, testLogger := setupTest()

			// Inject the test logger
			originalLogger := logger
			logger = testLogger
			defer func() { logger = originalLogger }()

			tt.logFunc(td.message)
			validateLogOutput(t, buf.String(), tt.level)
		})
	}

}

// TestFatal tests the Fatal function
func TestFatal(t *testing.T) {

	buf, testLogger := setupTest()

	// Inject test logger
	originalLogger := logger
	logger = testLogger
	defer func() { logger = originalLogger }()

	// Mock ExitFunc
	origExit := ExitFunc
	exitCode := 0
	ExitFunc = func(code int) { exitCode = code }
	defer func() { ExitFunc = origExit }()

	Fatal(td.message)

	// Fatal should exit with code 1
	if exitCode != 1 {
		t.Errorf("got exit code %d, want 1", exitCode)
	}

	// Check for Fatal level string (Magenta + [FTL])
	if !strings.Contains(buf.String(), "[FTL]") {
		t.Errorf("output missing fatal level tag")
	}

}

// TestEnabled tests the Enabled function behavior
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

			// We pass the level via Options, which CustomTextHandler respects
			h := NewCustomTextHandler(&bytes.Buffer{}, &slog.HandlerOptions{Level: tt.setLevel})

			if got := h.Enabled(context.Background(), tt.level); got != tt.want {
				t.Errorf("Enabled() = %v, want %v", got, tt.want)
			}

		})
	}

}

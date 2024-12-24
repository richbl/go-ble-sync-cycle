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

// Define test data
var td = testData{
	message:     "test message",
	level:       slog.LevelDebug,
	defaultOpts: slog.LevelDebug,
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

// TestNewCustomTextHandler tests the NewCustomTextHandler function
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

// TestCustomTextHandler tests the CustomTextHandler struct
func TestCustomTextHandler(t *testing.T) {

	// Define test cases
	tests := []testCase{
		{"debug", slog.LevelDebug, Blue + "[DBG]", 0},
		{"info", slog.LevelInfo, Green + "[INF]", 0},
		{"warn", slog.LevelWarn, Yellow + "[WRN]", 0},
		{"error", slog.LevelError, Magenta + "[ERR]", 0},
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

// TestLogLevels tests the log level functions
func TestLogLevels(t *testing.T) {

	// Define test cases
	tests := []struct {
		name    string
		logFunc func(interface{}, ...interface{})
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
			logger = testLogger

			tt.logFunc(td.message)
			validateLogOutput(t, buf.String(), tt.level)
		})
	}

}

// TestFatal tests the Fatal function
func TestFatal(t *testing.T) {

	buf, testLogger := setupTest()
	logger = testLogger

	origExit := ExitFunc
	exitCode := 0
	ExitFunc = func(code int) { exitCode = code }
	defer func() { ExitFunc = origExit }()

	Fatal(td.message)

	if exitCode != 0 {
		t.Errorf("got exit code %d, want 1", exitCode)
	}

	validateLogOutput(t, buf.String(), "FTL")
}

// TestEnabled tests the Enabled function
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

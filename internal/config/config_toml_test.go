package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSave verifies that the Save function correctly generates a TOML file
func TestSave(t *testing.T) {

	cfg := createTestConfig()
	version := "0.0.1-test"

	// Create temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "output_test.toml")

	// Execute Save
	if err := Save(tmpFile, cfg, version); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	// Read back raw content
	contentBytes, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read back saved file: %v", err)
	}

	content := string(contentBytes)

	// Sub-tests for specific validation logic
	t.Run("Version Header", func(t *testing.T) {

		if !strings.Contains(content, "# v"+version) {
			t.Errorf("Output missing version header. Got:\n%s", content)
		}

	})

	t.Run("String Values and Comments", func(t *testing.T) {

		// Expected: session_title = "Test Session"
		expectedLinePart := `session_title = "Test Session"`
		expectedComment := `# Short description`

		if !strings.Contains(content, expectedLinePart) {
			t.Errorf("Output missing config value '%s'", expectedLinePart)
		}

		if !strings.Contains(content, expectedComment) {
			t.Errorf("Output missing comment '%s'", expectedComment)
		}

	})

	t.Run("Numeric Formatting", func(t *testing.T) {

		// Check integer formatting
		if !strings.Contains(content, "scan_timeout_secs = 15") {
			t.Error("Output failed integer formatting check")
		}

		// Check float formatting (precision)
		if !strings.Contains(content, "speed_threshold = 0.50") {
			t.Error("Output failed float formatting check")
		}

	})

}

// TestPadToColumn verifies the helper function used to align comments
func TestPadToColumn(t *testing.T) {

	tests := []struct {
		input  string
		length int
	}{
		{"short", 40},
		{"a very long string that might exceed the column limit defined in the file", 2},
	}

	for _, tt := range tests {
		padding := padToColumn(tt.input)
		totalLen := len(tt.input) + len(padding)

		if len(tt.input) < commentColumn {

			if totalLen != commentColumn {
				t.Errorf("padToColumn(%q) resulted in total length %d, want %d", tt.input, totalLen, commentColumn)
			}

		} else {

			if len(padding) != 2 {
				t.Errorf("padToColumn(%q) (long string) padding len = %d, want 2", tt.input, len(padding))
			}

		}
	}

}

// createTestConfig returns a fully populated Config struct for testing
func createTestConfig() *Config {

	return &Config{
		App: AppConfig{
			SessionTitle: "Test Session",
			LogLevel:     "debug",
		},
		BLE: BLEConfig{
			SensorBDAddr:    "AA:BB:CC:DD:EE:FF",
			ScanTimeoutSecs: 15,
		},
		Speed: SpeedConfig{
			WheelCircumferenceMM: 2100,
			SpeedUnits:           "km/h",
			SpeedThreshold:       0.5,
			SmoothingWindow:      5,
		},
		Video: VideoConfig{
			MediaPlayer:       "mpv",
			FilePath:          "/tmp/video.mp4",
			SeekToPosition:    "00:00",
			WindowScaleFactor: 1.0,
			UpdateIntervalSec: 0.5,
			SpeedMultiplier:   1.0,
			OnScreenDisplay: VideoOSDConfig{
				DisplayCycleSpeed:    true,
				DisplayPlaybackSpeed: false,
				DisplayTimeRemaining: true,
				FontSize:             20,
				MarginX:              10,
				MarginY:              10,
			},
		},
	}

}

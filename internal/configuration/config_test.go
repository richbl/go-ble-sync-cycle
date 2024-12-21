package config

import (
	"fmt"
	"os"
	"testing"
)

// testData holds all test constants and configurations
type testData struct {
	filename     string
	invalidFile  string
	sensorUUID   string
	logLevel     string
	invalidLevel string
}

var td = testData{
	filename:     "test.mp4",
	invalidFile:  "non-existent-file.mp4",
	sensorUUID:   "test-uuid",
	logLevel:     logLevelDebug,
	invalidLevel: "invalid",
}

// testConfig represents a generic test configuration
type testConfig[T any] struct {
	name    string
	input   T
	wantErr bool
}

// generateConfigTOML returns valid or invalid TOML config based on isValid flag
func generateConfigTOML(isValid bool) string {
	// Generate valid and invalid TOML configs
	if isValid {
		return fmt.Sprintf(`
			[app]
			logging_level = "%s"

			[ble]
			sensor_uuid = "%s"
			scan_timeout_secs = 10

			[speed]
			smoothing_window = 5
			speed_threshold = 10.0
			wheel_circumference_mm = 2000
			speed_units = "%s"

			[video]
			file_path = "%s"
			display_playback_speed = true
			window_scale_factor = 1.0
			update_interval_sec = 1
			speed_multiplier = 1.0
		`, td.logLevel, td.sensorUUID, SpeedUnitsKMH, td.filename)
	}

	return fmt.Sprintf(`
		[app]
		logging_level = "%s"

		[ble]
		sensor_uuid = ""
		scan_timeout_secs = -1

		[speed]
		smoothing_window = -1
		speed_threshold = -10.0
		wheel_circumference_mm = -2000
		speed_units = "%s"

		[video]
		file_path = "%s"
		display_playback_speed = true
		window_scale_factor = -1.0
		update_interval_sec = -1
		speed_multiplier = -1.0
	`, td.invalidLevel, td.invalidLevel, td.invalidFile)
}

// createTempFile creates a temporary file with given content
func createTempFile(t *testing.T, prefix, content string) (string, func()) {
	t.Helper()

	// Create temp file
	tmpFile, err := os.CreateTemp("", prefix)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Write content to temp file
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}

	// Close temp file
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	// Return temp file path and cleanup
	return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
}

// runValidationTests runs validation tests for any config type
func runValidationTests[T any](t *testing.T, tests []testConfig[T]) {
	t.Helper()

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error

			if v, ok := any(tt.input).(interface{ validate() error }); ok {
				err = v.validate()
			} else if v, ok := any(&tt.input).(interface{ validate() error }); ok {
				err = v.validate()
			} else {
				t.Fatal("config does not implement validate() error")
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}

}

func TestLoadFile(t *testing.T) {
	// Create tests
	tests := []struct {
		name    string
		config  string
		wantErr bool
	}{
		{"valid config", generateConfigTOML(true), false},
		{"invalid config", generateConfigTOML(false), true},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, cleanup := createTempFile(t, "config", tt.config)
			defer cleanup()

			_, err := LoadFile(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFile() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}

}

// TestValidateAppConfig tests AppConfig validation
func TestValidateAppConfig(t *testing.T) {
	// Create tests
	tests := []testConfig[AppConfig]{
		{
			name:    "valid app config",
			input:   AppConfig{LogLevel: td.logLevel},
			wantErr: false,
		},
		{
			name:    "invalid log level",
			input:   AppConfig{LogLevel: td.invalidLevel},
			wantErr: true,
		},
	}

	// Run tests
	runValidationTests(t, tests)
}

// TestValidateBLEConfig tests BLEConfig validation
func TestValidateBLEConfig(t *testing.T) {
	// Create tests
	tests := []testConfig[BLEConfig]{
		{
			name: "valid BLE config",
			input: BLEConfig{
				SensorUUID:      td.sensorUUID,
				ScanTimeoutSecs: 10,
			},
			wantErr: false,
		},
		{
			name: "invalid BLE config",
			input: BLEConfig{
				SensorUUID:      "",
				ScanTimeoutSecs: -1,
			},
			wantErr: true,
		},
	}

	// Run tests
	runValidationTests(t, tests)
}

// TestValidateSpeedConfig tests SpeedConfig validation
func TestValidateSpeedConfig(t *testing.T) {
	// Create tests
	tests := []testConfig[SpeedConfig]{
		{
			name: "valid speed config",
			input: SpeedConfig{
				SmoothingWindow:      5,
				SpeedThreshold:       10.0,
				WheelCircumferenceMM: 2000,
				SpeedUnits:           SpeedUnitsKMH,
			},
			wantErr: false,
		},
		{
			name: "invalid speed config",
			input: SpeedConfig{
				SmoothingWindow:      -1,
				SpeedThreshold:       -10.0,
				WheelCircumferenceMM: -2000,
				SpeedUnits:           td.invalidLevel,
			},
			wantErr: true,
		},
	}

	// Run tests
	runValidationTests(t, tests)
}

// TestValidateVideoConfig tests VideoConfig validation
func TestValidateVideoConfig(t *testing.T) {

	// Create tests
	tests := []testConfig[VideoConfig]{
		{
			name: "valid video config",
			input: VideoConfig{
				FilePath: td.filename,
				OnScreenDisplay: VideoOSDConfig{
					DisplayPlaybackSpeed: true,
				},
				WindowScaleFactor: 1.0,
				UpdateIntervalSec: 1,
				SpeedMultiplier:   1.0,
			},
			wantErr: false,
		},
		{
			name: "invalid video config",
			input: VideoConfig{
				FilePath: td.invalidFile,
				OnScreenDisplay: VideoOSDConfig{
					DisplayPlaybackSpeed: true,
				},
				WindowScaleFactor: -1.0,
				UpdateIntervalSec: -1,
				SpeedMultiplier:   -1.0,
			},
			wantErr: true,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.input.FilePath != td.invalidFile {
				tmpFile, err := os.CreateTemp("", "test")

				if err != nil {
					t.Fatal(err)
				}

				defer os.Remove(tmpFile.Name())
				tt.input.FilePath = tmpFile.Name()
			}

			if err := tt.input.validate(); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}

}

package config

import (
	"os"
	"strings"
	"testing"
)

const (
	testFilename        = "test.mp4"
	invalidTestFilename = "non-existent-file.mp4"
	sensorTestUUID      = "test-uuid"
	logLevelInvalid     = "invalid"
)

// configTestCase is a helper struct for running validation tests
type configTestCase[T any] struct {
	name    string
	config  T
	wantErr bool
}

// Define TOML configurations for testing
var (
	validConfigTOML = `
		[app]
		logging_level = "` + logLevelDebug + `"

		[ble]
		sensor_uuid = "` + sensorTestUUID + `" 
		scan_timeout_secs = 10

		[speed]
		smoothing_window = 5
		speed_threshold = 10.0
		wheel_circumference_mm = 2000
		speed_units = "` + SpeedUnitsKMH + `"

		[video]
		file_path = "` + testFilename + `"
		display_playback_speed = true
		window_scale_factor = 1.0
		update_interval_sec = 1
		speed_multiplier = 1.0
	`

	invalidConfigTOML = `
		[app]
		logging_level = "` + logLevelInvalid + `"

		[ble]
		sensor_uuid = ""
		scan_timeout_secs = -1

		[speed]
		smoothing_window = -1
		speed_threshold = -10.0
		wheel_circumference_mm = -2000
		speed_units = "` + logLevelInvalid + `"

		[video]
		file_path = "` + invalidTestFilename + `"
		display_playback_speed = true
		window_scale_factor = -1.0
		update_interval_sec = -1
		speed_multiplier = -1.0
	`
)

// createTempConfigFile creates a temporary TOML configuration file
func createTempConfigFile(t *testing.T, config string) (string, func()) {

	// Create a configuration temporary file
	tmpFile, err := os.CreateTemp("", "config")
	if err != nil {
		t.Fatal(err)
	}

	// Replace the video file path with a temporary file
	if strings.Contains(config, testFilename) {

		// Create a temporary video file
		tmpVideoFile, err := os.CreateTemp("", "video")
		if err != nil {
			t.Fatal(err)
		}

		// Close the temporary video file
		if err := tmpVideoFile.Close(); err != nil {
			t.Fatal(err)
		}

		// Replace the video file path
		config = strings.ReplaceAll(config, testFilename, tmpVideoFile.Name())
		t.Cleanup(func() { os.Remove(tmpVideoFile.Name()) })
	}

	// Write the configuration to the temporary file
	if _, err := tmpFile.Write([]byte(config)); err != nil {
		t.Fatal(err)
	}

	// Close the temporary file
	if err := tmpFile.Close(); err != nil {
		t.Fatal(err)
	}

	return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }

}

// runValidationTest is a generic helper function for running validation tests
func runValidationTest[T any](t *testing.T, tests []configTestCase[T]) {

	// Run tests
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			// Handle both value and pointer receivers
			var err error

			if v, ok := any(tt.config).(interface{ validate() error }); ok {
				err = v.validate()
			} else if v, ok := any(&tt.config).(interface{ validate() error }); ok {
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

// TestLoadFile tests the LoadFile function
func TestLoadFile(t *testing.T) {

	// Define test cases
	tests := []configTestCase[string]{
		{name: "valid config", config: validConfigTOML, wantErr: false},
		{name: "invalid config", config: invalidConfigTOML, wantErr: true},
	}

	// Run tests
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			tmpFile, cleanup := createTempConfigFile(t, tt.config)
			defer cleanup()

			_, err := LoadFile(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFile() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}

}

// TestValidate tests the validate function
func TestValidate(t *testing.T) {

	// Define test cases
	tests := []configTestCase[string]{
		{name: "valid config", config: validConfigTOML, wantErr: false},
		{name: "invalid config", config: invalidConfigTOML, wantErr: true},
	}

	// Run tests
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			tmpFile, cleanup := createTempConfigFile(t, tt.config)
			defer cleanup()

			config, err := LoadFile(tmpFile)
			if err != nil {

				if !tt.wantErr {
					t.Errorf("LoadFile() error = %v, wantErr %v", err, tt.wantErr)
				}

				return

			}

			if err := config.validate(); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

// TestValidateVideoConfig tests the validate function
func TestValidateVideoConfig(t *testing.T) {

	// Define test cases
	tests := []configTestCase[VideoConfig]{
		{
			name: "valid config",
			config: VideoConfig{
				FilePath: testFilename,
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
			name: "invalid config",
			config: VideoConfig{
				FilePath: invalidTestFilename,
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

			if tt.config.FilePath != invalidTestFilename {
				tmpFile, err := os.CreateTemp("", "test")

				if err != nil {
					t.Fatal(err)
				}

				defer os.Remove(tmpFile.Name())
				tt.config.FilePath = tmpFile.Name()

			}

			if err := tt.config.validate(); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}

}

// TestValidateAppConfig tests the validate function
func TestValidateAppConfig(t *testing.T) {

	tests := []configTestCase[AppConfig]{
		{
			name:    "valid config",
			config:  AppConfig{LogLevel: logLevelDebug},
			wantErr: false,
		},
		{
			name:    "invalid config",
			config:  AppConfig{LogLevel: logLevelInvalid},
			wantErr: true,
		},
	}

	runValidationTest(t, tests)

}

// TestValidateBLEConfig tests the validate function
func TestValidateBLEConfig(t *testing.T) {

	tests := []configTestCase[BLEConfig]{
		{
			name: "valid config",
			config: BLEConfig{
				SensorUUID:      sensorTestUUID,
				ScanTimeoutSecs: 10,
			},
			wantErr: false,
		},
		{
			name: "invalid config",
			config: BLEConfig{
				SensorUUID:      "",
				ScanTimeoutSecs: -1,
			},
			wantErr: true,
		},
	}

	runValidationTest(t, tests)

}

// TestValidateSpeedConfig tests the validate function
func TestValidateSpeedConfig(t *testing.T) {

	tests := []configTestCase[SpeedConfig]{
		{
			name: "valid config",
			config: SpeedConfig{
				SmoothingWindow:      5,
				SpeedThreshold:       10.0,
				WheelCircumferenceMM: 2000,
				SpeedUnits:           SpeedUnitsKMH,
			},
			wantErr: false,
		},
		{
			name: "invalid config",
			config: SpeedConfig{
				SmoothingWindow:      -1,
				SpeedThreshold:       -10.0,
				WheelCircumferenceMM: -2000,
				SpeedUnits:           logLevelInvalid,
			},
			wantErr: true,
		},
	}

	runValidationTest(t, tests)

}

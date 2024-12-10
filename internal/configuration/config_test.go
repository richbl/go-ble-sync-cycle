package config

import (
	"os"
	"strings"
	"testing"
)

// createTempConfigFile creates temporary TOML config file for testing
func createTempConfigFile(t *testing.T, config string) (string, func()) {

	// Create a temporary config file
	tmpFile, err := os.CreateTemp("", "config")
	if err != nil {
		t.Fatal(err)
	}

	// If "test.mp4" identified in the config, replace it with a temporary video file
	if strings.Contains(config, "test.mp4") {

		tmpVideoFile, err := os.CreateTemp("", "video")
		if err != nil {
			t.Fatal(err)
		}

		if err := tmpVideoFile.Close(); err != nil {
			t.Fatal(err)
		}

		config = strings.ReplaceAll(config, "test.mp4", tmpVideoFile.Name())
		t.Cleanup(func() {

			os.Remove(tmpVideoFile.Name())
		})

	}

	// Write the TOML configuration to the temporary file
	if _, err := tmpFile.Write([]byte(config)); err != nil {
		t.Fatal(err)
	}

	if err := tmpFile.Close(); err != nil {
		t.Fatal(err)
	}

	// Return the file path and a deferred function for cleaning up
	return tmpFile.Name(), func() {
		os.Remove(tmpFile.Name())
	}

}

// TestLoadFile tests the LoadFile() function
func TestLoadFile(t *testing.T) {

	// Define test cases
	tests := []struct {
		name    string
		config  string
		wantErr bool
	}{
		{
			name: "valid config",
			config: `
				[app]
				logging_level = "debug"

				[ble]
				sensor_uuid = "some-uuid"
				scan_timeout_secs = 10

				[speed]
				smoothing_window = 5
				speed_threshold = 10.0
				wheel_circumference_mm = 2000
				speed_units = "km/h"

				[video]
				file_path = "test.mp4"
				display_playback_speed = true
				window_scale_factor = 1.0
				update_interval_sec = 1
				speed_multiplier = 1.0
					`,
			wantErr: false,
		},
		{
			name: "invalid config",
			config: `
				[app]
				logging_level = "invalid"

				[ble]
				sensor_uuid = ""
				scan_timeout_secs = -1

				[speed]
				smoothing_window = -1
				speed_threshold = -10.0
				wheel_circumference_mm = -2000
				speed_units = "invalid"

				[video]
				file_path = "non-existent-file.mp4"
				display_playback_speed = true
				window_scale_factor = -1.0
				update_interval_sec = -1
				speed_multiplier = -1.0
					`,
			wantErr: true,
		},
	}

	// Loop through the test cases
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			tmpFile, cleanup := createTempConfigFile(t, tt.config)
			defer cleanup()

			// Call the LoadFile() function
			_, err := LoadFile(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFile() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}

}

// TestValidate tests the validate() function
func TestValidate(t *testing.T) {

	// Define test cases
	tests := []struct {
		name    string
		config  string
		wantErr bool
	}{
		{
			name: "valid config",
			config: `
				[app]
				logging_level = "debug"

				[ble]
				sensor_uuid = "some-uuid"
				scan_timeout_secs = 10

				[speed]
				smoothing_window = 5
				speed_threshold = 10.0
				wheel_circumference_mm = 2000
				speed_units = "km/h"

				[video]
				file_path = "test.mp4"
				display_playback_speed = true
				window_scale_factor = 1.0
				update_interval_sec = 1
				speed_multiplier = 1.0
					`,
			wantErr: false,
		},
		{
			name: "invalid config",
			config: `
				[app]
				logging_level = "invalid"

				[ble]
				sensor_uuid = ""
				scan_timeout_secs = -1

				[speed]
				smoothing_window = -1
				speed_threshold = -10.0
				wheel_circumference_mm = -2000
				speed_units = "invalid"

				[video]
				file_path = "non-existent-file.mp4"
				display_playback_speed = true
				window_scale_factor = -1.0
				update_interval_sec = -1
				speed_multiplier = -1.0
					`,
			wantErr: true,
		},
	}

	// Loop through the test cases
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			tmpFile, cleanup := createTempConfigFile(t, tt.config)
			defer cleanup()

			// Call the LoadFile() function
			_, err := LoadFile(tmpFile)
			if err != nil && !tt.wantErr {
				t.Errorf("LoadFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			config, err := LoadFile(tmpFile)
			if err != nil {
				t.Fatal(err)
			}

			if err := config.validate(); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}

}

// TestValidateVideoConfig tests the validate() function
func TestValidateVideoConfig(t *testing.T) {

	// Define test cases
	tests := []struct {
		name    string
		config  VideoConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: VideoConfig{
				FilePath:             "test.mp4",
				DisplayPlaybackSpeed: true,
				WindowScaleFactor:    1.0,
				UpdateIntervalSec:    1,
				SpeedMultiplier:      1.0,
			},
			wantErr: false,
		},
		{
			name: "invalid config",
			config: VideoConfig{
				FilePath:             "non-existent-file.mp4",
				DisplayPlaybackSpeed: true,
				WindowScaleFactor:    -1.0,
				UpdateIntervalSec:    -1,
				SpeedMultiplier:      -1.0,
			},
			wantErr: true,
		},
	}

	// Loop through the test cases
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			if tt.config.FilePath != "non-existent-file.mp4" {

				tmpFile, err := os.CreateTemp("", "test")
				if err != nil {
					t.Fatal(err)
				}

				defer os.Remove(tmpFile.Name())

				// Write the TOML configuration to the temporary file and validate it
				tt.config.FilePath = tmpFile.Name()
			}

			if err := tt.config.validate(); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}

}

// TestValidateAppConfig tests the validate() function
func TestValidateAppConfig(t *testing.T) {

	// Define test cases
	tests := []struct {
		name    string
		config  AppConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: AppConfig{
				LogLevel: logLevelDebug,
			},
			wantErr: false,
		},
		{
			name: "invalid config",
			config: AppConfig{
				LogLevel: "invalid",
			},
			wantErr: true,
		},
	}

	// Loop through the test cases
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			if err := tt.config.validate(); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}

}

// TestValidateBLEConfig tests the validate() function
func TestValidateBLEConfig(t *testing.T) {

	// Define test cases
	tests := []struct {
		name    string
		config  BLEConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: BLEConfig{
				SensorUUID:      "some-uuid",
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

	// Loop through the test cases
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			if err := tt.config.validate(); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}

}

// TestValidateSpeedConfig tests the validate() function
func TestValidateSpeedConfig(t *testing.T) {

	// Define test cases
	tests := []struct {
		name    string
		config  SpeedConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: SpeedConfig{
				SmoothingWindow:      5,
				SpeedThreshold:       10.0,
				WheelCircumferenceMM: 2000,
				SpeedUnits:           speedUnitsKMH,
			},
			wantErr: false,
		},
		{
			name: "invalid config",
			config: SpeedConfig{
				SmoothingWindow:      -1,
				SpeedThreshold:       -10.0,
				WheelCircumferenceMM: -2000,
				SpeedUnits:           "invalid",
			},
			wantErr: true,
		},
	}

	// Loop through the test cases
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			if err := tt.config.validate(); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}

}

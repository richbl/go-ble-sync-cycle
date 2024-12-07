package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
)

// TestLoadFile tests various scenarios for the LoadFile function
func TestLoadFile(t *testing.T) {

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				App: AppConfig{
					LogLevel: "info",
				},
				BLE: BLEConfig{
					SensorUUID:      "some-uuid",
					ScanTimeoutSecs: 10,
				},
				Speed: SpeedConfig{
					SmoothingWindow:      5,
					SpeedThreshold:       10.0,
					WheelCircumferenceMM: 2000,
					SpeedUnits:           "km/h",
				},
				Video: VideoConfig{
					FilePath:          "test.mp4",
					DisplaySpeed:      true,
					WindowScaleFactor: 1.5,
					UpdateIntervalSec: 2,
					SpeedMultiplier:   1.2,
				},
			},
			wantErr: false,
		},
		//... other test cases no covered by the other validation tests...
	}

	// Run tests for each test scenario
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			// Create a temporary directory
			tmpDir, err := os.MkdirTemp("", "config-test")
			if err != nil {
				t.Fatal(err)
			}

			// Clean up the temporary directory
			defer os.RemoveAll(tmpDir)

			// Create a temporary config file
			tmpFile, err := os.CreateTemp(tmpDir, "config-")
			if err != nil {
				t.Fatal(err)
			}

			// Clean up the temporary config file
			defer os.Remove(tmpFile.Name())

			// Create a temporary video file
			videoFile, err := os.Create(filepath.Join(tmpDir, "test.mp4"))
			if err != nil {
				t.Fatal(err)
			}

			// Clean up the temporary video file
			defer os.Remove(videoFile.Name())
			err = videoFile.Close()

			if err != nil {
				t.Fatal(err)
			}

			// Update the configuration to include the full path to the video file
			tt.config.Video.FilePath = filepath.Join(tmpDir, "test.mp4")

			var buf bytes.Buffer

			enc := toml.NewEncoder(&buf)
			if err := enc.Encode(tt.config); err != nil {
				t.Fatal(err)
			}

			_, err = tmpFile.WriteString(buf.String())
			if err != nil {
				t.Fatal(err)
			}

			err = tmpFile.Close()
			if err != nil {
				t.Fatal(err)
			}

			_, err = LoadFile(tmpFile.Name())
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFile() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}

}

// TestValidateLogLevel tests the validateLogLevel function
func TestValidateLogLevel(t *testing.T) {

	// Define test cases
	tests := []struct {
		name     string
		logLevel string
		wantErr  bool
	}{
		{
			name:     "valid log level",
			logLevel: "info",
			wantErr:  false,
		}, {
			name:     "invalid log level",
			logLevel: "invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			ac := &AppConfig{LogLevel: tt.logLevel}

			if err := ac.validateLogLevel(); (err != nil) != tt.wantErr {
				t.Errorf("validateLogLevel() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}

}

// TestValidateSpeedUnits tests the validateSpeedUnits function
func TestValidateSpeedUnits(t *testing.T) {

	// Define test cases
	tests := []struct {
		name       string
		speedUnits string
		wantErr    bool
	}{
		{
			name:       "valid speed units",
			speedUnits: "km/h",
			wantErr:    false,
		}, {
			name:       "invalid speed units",
			speedUnits: "invalid",
			wantErr:    true,
		},
	}

	// Run tests
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			sc := &SpeedConfig{SpeedUnits: tt.speedUnits}
			if err := sc.validateSpeedUnits(); (err != nil) != tt.wantErr {
				t.Errorf("validateSpeedUnits() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}

}

// TestValidateVideoFile tests the validateVideoFile function
func TestValidateVideoFile(t *testing.T) {

	// Define test cases
	tests := []struct {
		name     string
		filePath string
		wantErr  bool
	}{
		{
			name:     "existing video file",
			filePath: "./config_test.go", // This file exists, so we use it as an example
			wantErr:  false,
		}, {
			name:     "non-existent video file",
			filePath: "non-existent.mp4",
			wantErr:  true,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			vc := &VideoConfig{FilePath: tt.filePath}
			if err := vc.validateVideoFile(); (err != nil) != tt.wantErr {
				t.Errorf("validateVideoFile() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}

}

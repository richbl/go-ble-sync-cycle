package config

import (
	"testing"
)

const (
	testVideo = "test_video.mp4"
)

// TestLoad tests the Load function
func TestLoad(t *testing.T) {

	// Define test cases
	tests := []struct {
		name        string
		configFile  string
		expectError bool
	}{
		{
			name:        "valid config file",
			configFile:  "config_test.toml",
			expectError: false,
		},
		{
			name:        "invalid config file",
			configFile:  "invalid_config.toml",
			expectError: true,
		},
		{
			name:        "non-existent config file",
			configFile:  "non_existent.toml",
			expectError: true,
		},
	}

	// Run tests
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			_, err := Load(tt.configFile)
			if (err != nil) != tt.expectError {
				t.Errorf("Load() error = %v, expectError %v", err, tt.expectError)
			}

		})
	}

}

// TestAppConfigValidate tests the AppConfig validate function
func TestAppConfigValidate(t *testing.T) {

	// Define test cases
	tests := []struct {
		name        string
		logLevel    string
		expectError bool
	}{
		{"valid debug", logLevelDebug, false},
		{"valid info", logLevelInfo, false},
		{"valid warn", logLevelWarn, false},
		{"valid error", logLevelError, false},
		{"valid fatal", logLevelFatal, false},
		{"invalid log level", "invalid", true},
	}

	// Run tests
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			ac := AppConfig{LogLevel: tt.logLevel}
			err := ac.validate()
			if (err != nil) != tt.expectError {
				t.Errorf("AppConfig.validate() error = %v, expectError %v", err, tt.expectError)
			}

		})
	}

}

// TestBLEConfigValidate tests the BLEConfig validate function
func TestBLEConfigValidate(t *testing.T) {

	// Define test cases
	tests := []struct {
		name            string
		sensorBDAddr    string
		scanTimeoutSecs int
		expectError     bool
	}{
		{"valid BD_ADDR and timeout", "00:11:22:33:44:55", 10, false},
		{"invalid BD_ADDR", "invalid", 10, true},
		{"invalid scan timeout", "00:11:22:33:44:55", 0, true},
	}

	// Run tests
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			bc := BLEConfig{SensorBDAddr: tt.sensorBDAddr, ScanTimeoutSecs: tt.scanTimeoutSecs}
			err := bc.validate()
			if (err != nil) != tt.expectError {
				t.Errorf("BLEConfig.validate() error = %v, expectError %v", err, tt.expectError)
			}

		})
	}

}

// TestSpeedConfigValidate tests the SpeedConfig validate function
func TestSpeedConfigValidate(t *testing.T) {

	// Define test cases
	tests := []struct {
		name               string
		smoothingWindow    int
		speedThreshold     float64
		wheelCircumference int
		speedUnits         string
		expectError        bool
	}{
		{"valid config", 10, 5.0, 1000, SpeedUnitsKMH, false},
		{"invalid smoothing window", 0, 5.0, 1000, SpeedUnitsKMH, true},
		{"invalid speed threshold", 10, 11.0, 1000, SpeedUnitsKMH, true},
		{"invalid wheel circumference", 10, 5.0, 49, SpeedUnitsKMH, true},
		{"invalid speed units", 10, 5.0, 1000, "invalid", true},
	}

	// Run tests
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			sc := SpeedConfig{
				SmoothingWindow:      tt.smoothingWindow,
				SpeedThreshold:       tt.speedThreshold,
				WheelCircumferenceMM: tt.wheelCircumference,
				SpeedUnits:           tt.speedUnits,
			}

			err := sc.validate()
			if (err != nil) != tt.expectError {
				t.Errorf("SpeedConfig.validate() error = %v, expectError %v", err, tt.expectError)
			}

		})
	}

}

// TestVideoConfigValidate tests the VideoConfig validate function
func TestVideoConfigValidate(t *testing.T) {

	// Define test cases
	tests := []struct {
		name              string
		mediaPlayer       string
		filePath          string
		windowScaleFactor float64
		seekToPosition    string
		updateIntervalSec float64
		speedMultiplier   float64
		fontSize          int
		expectError       bool
	}{
		{"valid config", MediaPlayerMPV, testVideo, 0.5, "00:30", 1.0, 0.5, 20, false},
		{"invalid media player", "xyz", testVideo, 0.5, "00:30", 1.0, 0.5, 20, true},
		{"invalid file path", MediaPlayerMPV, "invalid_path.mp4", 0.5, "00:30", 1.0, 0.5, 20, true},
		{"invalid window scale factor", MediaPlayerMPV, testVideo, 1.1, "00:30", 1.0, 0.5, 20, true},
		{"invalid seek position", MediaPlayerMPV, testVideo, 0.5, "invalid", 1.0, 0.5, 20, true},
		{"invalid update interval", MediaPlayerMPV, testVideo, 0.5, "00:30", 3.1, 0.5, 20, true},
		{"invalid speed multiplier", MediaPlayerMPV, testVideo, 0.5, "00:30", 1.0, 1.6, 20, true},
		{"invalid font size", MediaPlayerMPV, testVideo, 0.5, "00:30", 1.0, 0.5, 201, true},
	}

	// Run tests
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			vc := VideoConfig{
				MediaPlayer:       tt.mediaPlayer,
				FilePath:          tt.filePath,
				WindowScaleFactor: tt.windowScaleFactor,
				SeekToPosition:    tt.seekToPosition,
				UpdateIntervalSec: tt.updateIntervalSec,
				SpeedMultiplier:   tt.speedMultiplier,
				OnScreenDisplay: VideoOSDConfig{
					FontSize: tt.fontSize,
				},
			}

			err := vc.validate()
			if (err != nil) != tt.expectError {
				t.Errorf("VideoConfig.validate() error = %v, expectError %v", err, tt.expectError)
			}

		})
	}

}

// TestVideoOSDConfigValidate tests the VideoOSDConfig validate function
func TestValidateTimeFormat(t *testing.T) {

	// Define test cases
	tests := []struct {
		name        string
		input       string
		expectValid bool
	}{
		{"valid MM:SS", "01:30", true},
		{"valid SS", "90", true},
		{"invalid MM:SS", "01:60", false},
		{"invalid SS", "-10", false},
		{"invalid format", "invalid", false},
	}

	// Run tests
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			valid := validateTimeFormat(tt.input)
			if valid != tt.expectValid {
				t.Errorf("validateTimeFormat() = %v, expectValid %v", valid, tt.expectValid)
			}

		})
	}

}

// TestVideoOSDConfigValidate tests the VideoOSDConfig validate function
func TestValidateField(t *testing.T) {

	// Define test cases
	tests := []struct {
		name        string
		value       any
		min         any
		max         any
		expectError bool
	}{
		{"valid int", 10, 1, 20, false},
		{"invalid int", 0, 1, 20, true},
		{"valid float64", 1.5, 1.0, 2.0, false},
		{"invalid float64", 0.5, 1.0, 2.0, true},
		{"unsupported type", "invalid", 1, 20, true},
	}

	// Run tests
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			err := validateField(tt.value, tt.min, tt.max, errInvalidLogLevel)
			if (err != nil) != tt.expectError {
				t.Errorf("validateField() error = %v, expectError %v", err, tt.expectError)
			}

		})
	}

}

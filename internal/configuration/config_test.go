package config

import (
	"os"
	"testing"
)

// TestAppConfigValidate tests the validation of the AppConfig
func TestAppConfigValidate(t *testing.T) {

	// Define test cases
	tests := []struct {
		name     string
		logLevel string
		wantErr  bool
	}{
		{"Valid Debug", logLevelDebug, false},
		{"Valid Info", logLevelInfo, false},
		{"Valid Warn", logLevelWarn, false},
		{"Valid Error", logLevelError, false},
		{"Valid Fatal", logLevelFatal, false},
		{"Invalid Log Level", "invalid", true},
	}

	// Run tests
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			ac := &AppConfig{LogLevel: tt.logLevel}

			if err := ac.validate(); (err != nil) != tt.wantErr {
				t.Errorf("AppConfig.validate() error = %v, wantErr %v", err, tt.wantErr)
			}

		})

	}

}

// TestBLEConfigValidate tests the validation of the BLEConfig
func TestBLEConfigValidate(t *testing.T) {

	// Define test cases
	tests := []struct {
		name       string
		sensorUUID string
		wantErr    bool
	}{
		{"Valid UUID", "123e4567-e89b-12d3-a456-426614174000", false},
		{"Empty UUID", "", true},
	}

	// Run tests
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			bc := &BLEConfig{SensorUUID: tt.sensorUUID}

			if err := bc.validate(); (err != nil) != tt.wantErr {
				t.Errorf("BLEConfig.validate() error = %v, wantErr %v", err, tt.wantErr)
			}

		})

	}

}

// TestSpeedConfigValidate tests the validation of the SpeedConfig
func TestSpeedConfigValidate(t *testing.T) {

	// Define test cases
	tests := []struct {
		name       string
		speedUnits string
		wantErr    bool
	}{
		{"Valid KMH", SpeedUnitsKMH, false},
		{"Valid MPH", SpeedUnitsMPH, false},
		{"Invalid Units", "invalid", true},
	}

	// Run tests
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			sc := &SpeedConfig{SpeedUnits: tt.speedUnits}

			if err := sc.validate(); (err != nil) != tt.wantErr {
				t.Errorf("SpeedConfig.validate() error = %v, wantErr %v", err, tt.wantErr)
			}

		})

	}

}

// TestVideoConfigValidate tests the validation of the VideoConfig
func TestVideoConfigValidate(t *testing.T) {

	// Create a temporary file for testing
	tmpFile := createTempFile(t)
	defer os.Remove(tmpFile.Name())

	// Define test cases
	tests := []struct {
		name              string
		filePath          string
		updateIntervalSec float64
		SeekToPosition    string
		wantErr           bool
	}{
		{"Valid File, Interval and Seek", tmpFile.Name(), 1.0, "1:00", false},
		{"Invalid File", "whatfile?", 1.0, "1:00", true},
		{"Invalid Interval", tmpFile.Name(), 0.0, "1:00", true},
		{"Invalid Seek", tmpFile.Name(), 1.0, "invalid", true},
	}

	// Run tests
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			vc := &VideoConfig{
				FilePath:          tt.filePath,
				UpdateIntervalSec: tt.updateIntervalSec,
				SeekToPosition:    tt.SeekToPosition,
			}

			if err := vc.validate(); (err != nil) != tt.wantErr {
				t.Errorf("VideoConfig.validate() error = %v, wantErr %v", err, tt.wantErr)
			}

		})

	}

}

// TestConfigValidate tests the validation of the Config
func TestConfigValidate(t *testing.T) {

	// Create a temporary file for testing
	tmpFile := createTempFile(t)
	defer os.Remove(tmpFile.Name())

	// Define test cases
	validConfig := &Config{
		App:   AppConfig{LogLevel: logLevelInfo},
		BLE:   BLEConfig{SensorUUID: "123e4567-e89b-12d3-a456-426614174000"},
		Speed: SpeedConfig{SpeedUnits: SpeedUnitsKMH},
		Video: VideoConfig{
			FilePath:          tmpFile.Name(), // Use the temporary file path
			SeekToPosition:    "1:00",
			UpdateIntervalSec: 1.0,
		},
	}

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{"Valid Config", validConfig, false},
		{"Invalid App Config", &Config{App: AppConfig{LogLevel: "invalid"}}, true},
		{"Invalid BLE Config", &Config{BLE: BLEConfig{SensorUUID: ""}}, true},
		{"Invalid Speed Config", &Config{Speed: SpeedConfig{SpeedUnits: "invalid"}}, true},
		{"Invalid Video Config", &Config{Video: VideoConfig{UpdateIntervalSec: 0.0}}, true},
	}

	// Run tests
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			if err := tt.config.validate(); (err != nil) != tt.wantErr {
				t.Errorf("Config.validate() error = %v, wantErr %v", err, tt.wantErr)
			}

		})

	}

}

// createTempFile creates a temporary file for testing
func createTempFile(t *testing.T) os.File {

	tmpFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatal(err)
	}

	return *tmpFile
}

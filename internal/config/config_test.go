package config

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"
)

// TestLoad tests the Load function
func TestLoad(t *testing.T) {

	// Define test cases (errIs is optional, and only checked when the error chain allows it)
	tests := []struct {
		errIs       error
		name        string
		configFile  string
		expectError bool
	}{
		{
			name:       "valid config file",
			configFile: writeTestConfig(t, "valid.toml"),
		},
		{
			name:        "malformed config file",
			configFile:  writeTestFile(t, "malformed.toml", malformedTOML),
			expectError: true,
		},
		{
			name:        "config file with missing video file",
			configFile:  writeTestConfig(t, "missing_video.toml", withMissingVideoFile),
			expectError: true,
			errIs:       fs.ErrNotExist,
		},
		{
			name:        "non-existent config file",
			configFile:  filepath.Join(t.TempDir(), "non_existent.toml"),
			expectError: true,
			errIs:       fs.ErrNotExist,
		},
	}

	// Run tests
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			cfg, err := Load(tt.configFile)
			if (err != nil) != tt.expectError {
				t.Fatalf("Load() error = %v, expectError %v", err, tt.expectError)
			}

			if tt.errIs != nil && !errors.Is(err, tt.errIs) {
				t.Errorf("Load() error = %v, want error matching %v", err, tt.errIs)
			}

			if !tt.expectError && cfg == nil {
				t.Error("Load() returned a nil config without an error")
			}

		})
	}

}

// TestAppConfigValidate tests the AppConfig validate function
func TestAppConfigValidate(t *testing.T) {

	sessionTitle := "a valid title"

	// Define test cases
	tests := []struct {
		name         string
		logLevel     string
		sessionTitle string
		expectError  bool
	}{
		{"valid debug", logLevelDebug, sessionTitle, false},
		{"valid info", logLevelInfo, sessionTitle, false},
		{"valid warn", logLevelWarn, sessionTitle, false},
		{"valid error", logLevelError, sessionTitle, false},
		{"valid fatal", logLevelFatal, sessionTitle, false},
		{"invalid log level", "invalid", sessionTitle, true},
		{"valid session title", logLevelInfo, sessionTitle, false},
		{"invalid session title", logLevelInfo, "This is a very long session title that is designed to be well over the two hundred character limit that has been imposed on it to ensure that the validation logic is correctly catching strings that are too long.", true},
	}

	// Run tests
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			ac := AppConfig{
				LogLevel:     tt.logLevel,
				SessionTitle: tt.sessionTitle,
			}
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
		{"invalid speed units", 10, 5.0, 1000, "invalid", true},
		{"invalid smoothing window", 0, 5.0, 1000, SpeedUnitsKMH, true},
		{"invalid speed threshold", 10, 11.0, 1000, SpeedUnitsKMH, true},
		{"invalid wheel circumference", 10, 5.0, 49, SpeedUnitsKMH, true},
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

	// Define test cases: each one breaks a single setting of an otherwise valid VideoConfig, and
	// names the error that setting must produce (nil means the config is valid), so a case can't
	// pass because of some unrelated failure
	tests := []struct {
		mutate  func(vc *VideoConfig)
		wantErr error
		name    string
	}{
		{func(*VideoConfig) {}, nil, "valid config"},
		{func(vc *VideoConfig) { vc.MediaPlayer = "xyz" }, errInvalidPlayer, "invalid media player"},
		{func(vc *VideoConfig) { vc.FilePath = filepath.Join(filepath.Dir(vc.FilePath), "missing.mp4") }, fs.ErrNotExist, "missing video file"},
		{func(vc *VideoConfig) { vc.WindowScaleFactor = 1.1 }, errWindowScale, "invalid window scale factor"},
		{func(vc *VideoConfig) { vc.SeekToPosition = "invalid" }, errInvalidSeek, "invalid seek position"},
		{func(vc *VideoConfig) { vc.UpdateIntervalSec = 3.1 }, errInvalidInterval, "invalid update interval"},
		{func(vc *VideoConfig) { vc.SpeedMultiplier = 1.6 }, errSpeedMultiplier, "invalid speed multiplier"},
		{func(vc *VideoConfig) { vc.OnScreenDisplay.FontSize = 201 }, errFontSize, "invalid font size"},
		{func(vc *VideoConfig) { vc.OnScreenDisplay.AlignX = "invalid" }, errInvalidAlignX, "invalid OSD align x"},
		{func(vc *VideoConfig) { vc.OnScreenDisplay.AlignY = "invalid" }, errInvalidAlignY, "invalid OSD align y"},
		{func(vc *VideoConfig) { vc.OnScreenDisplay.MarginX = 301 }, errOSDMargin, "invalid OSD margin x"},
		{func(vc *VideoConfig) { vc.OnScreenDisplay.MarginY = 601 }, errOSDMargin, "invalid OSD margin y"},
	}

	// Run tests
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			vc := NewDefault(newTestVideoFile(t)).Video
			tt.mutate(&vc)

			// errors.Is(nil, nil) is true, so this covers the valid case as well
			if err := vc.validate(); !errors.Is(err, tt.wantErr) {
				t.Errorf("VideoConfig.validate() error = %v, want %v", err, tt.wantErr)
			}

		})
	}

}

// TestValidateTimeFormat tests the validateTimeFormat function
func TestValidateTimeFormat(t *testing.T) {

	// Define test cases
	tests := []struct {
		name        string
		input       string
		expectValid bool
	}{
		{"valid HH:MM:SS", "00:01:30", true},
		{"valid maximums", "99:59:59", true},
		{"invalid MM:SS", "01:30", false},
		{"invalid SS", "90", false},
		{"invalid hours", "100:00:00", false},
		{"invalid minutes", "00:60:00", false},
		{"invalid seconds", "00:00:60", false},
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

// TestValidateField tests the validateField function
func TestValidateField(t *testing.T) {

	// Define test cases
	tests := []struct {
		name        string
		value       any
		min         any
		max         any
		expectError bool
	}{
		{"invalid int", 0, 1, 20, true},
		{"valid float64", 1.5, 1.0, 2.0, false},
		{"invalid float64", 0.5, 1.0, 2.0, true},
		{"valid int", 10, 1, 20, false},
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

package config

import (
	"errors"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config represents the application configuration
type Config struct {
	App   AppConfig   `toml:"app"`
	BLE   BLEConfig   `toml:"ble"`
	Speed SpeedConfig `toml:"speed"`
	Video VideoConfig `toml:"video"`
}

// AppConfig represents the application configuration
type AppConfig struct {
	LogLevel string `toml:"logging_level"`
}

// BLEConfig represents the BLE controller configuration
type BLEConfig struct {
	SensorUUID      string `toml:"sensor_uuid"`
	ScanTimeoutSecs int    `toml:"scan_timeout_secs"`
}

// SpeedConfig represents the speed controller configuration
type SpeedConfig struct {
	SmoothingWindow      int     `toml:"smoothing_window"`
	SpeedThreshold       float64 `toml:"speed_threshold"`
	WheelCircumferenceMM int     `toml:"wheel_circumference_mm"`
	SpeedUnits           string  `toml:"speed_units"`
}

// VideoConfig represents the MPV video player configuration
type VideoConfig struct {
	FilePath          string  `toml:"file_path"`
	DisplaySpeed      bool    `toml:"display_speed"`
	WindowScaleFactor float64 `toml:"window_scale_factor"`
	UpdateIntervalSec int     `toml:"update_interval_sec"`
	SpeedMultiplier   float64 `toml:"speed_multiplier"`
}

// Set valid configuration values in TOML
var validLogLevels = []string{"debug", "info", "warn", "error"}
var validSpeedUnits = []string{"km/h", "mph"}

// LoadFile loads the application configuration from the given filepath
func LoadFile(filepath string) (*Config, error) {

	var config Config

	// Load configuration file (TOML)
	_, err := toml.DecodeFile(filepath, &config)
	if err != nil {
		return nil, err
	}

	// Validate user logging level provided in TOML file
	if err := config.App.validateLogLevel(); err != nil {
		return nil, err
	}

	// Validate user speed units provided in TOML file
	if err := config.Speed.validateSpeedUnits(); err != nil {
		return nil, err
	}

	// Validate sensor UUID provided in TOML file
	if config.BLE.SensorUUID == "" {
		return nil, errors.New("sensor UUID must be specified in configuration")
	}

	// Validate video file path provided in TOML file
	if err := config.Video.validateVideoFile(); err != nil {
		return nil, err
	}

	return &config, nil

}

// ValidateLogLevel validates the log level provided in the TOML file
func (ac *AppConfig) validateLogLevel() error {

	if !contains(validLogLevels, ac.LogLevel) {
		validLevelsStr := strings.Join(validLogLevels, ", ")
		return errors.New("invalid log level " + ac.LogLevel + ": must be one of [" + validLevelsStr + "]")
	}

	return nil

}

// ValidateSpeedUnits validates the speed units provided in the TOML file
func (sc *SpeedConfig) validateSpeedUnits() error {

	if !contains(validSpeedUnits, sc.SpeedUnits) {
		validSpeedUnitsStr := strings.Join(validSpeedUnits, ", ")
		return errors.New("invalid speed units " + sc.SpeedUnits + ": must be one of [" + validSpeedUnitsStr + "]")
	}

	return nil

}

// ValidateVideoFile validates the video file path provided in the TOML file
func (vc *VideoConfig) validateVideoFile() error {

	if _, err := os.Stat(vc.FilePath); os.IsNotExist(err) {
		return errors.New("invalid video file " + vc.FilePath + ": file does not exist")
	}

	return nil

}

// contains returns true if the string list contains the specified string
func contains(list []string, str string) bool {

	for _, v := range list {

		if v == str {
			return true
		}

	}

	return false

}

package config

import (
	"errors"

	"github.com/BurntSushi/toml"
)

// Config represents the application configuration
type Config struct {
	BLE   BLEConfig   `toml:"ble"`
	Speed SpeedConfig `toml:"speed"`
	Video VideoConfig `toml:"video"`
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
	UpdateIntervalSec int     `toml:"update_interval_sec"`
	SpeedMultiplier   float64 `toml:"speed_multiplier"`
}

// LoadFile loads the application configuration from the given filepath
func LoadFile(filepath string) (*Config, error) {

	var config Config

	// Load configuration file (TOML)
	_, err := toml.DecodeFile(filepath, &config)
	if err != nil {
		return nil, err
	}

	// Validate video file path provided in TOML file
	if config.Video.FilePath == "" {
		return nil, errors.New("video file path must be specified in configuration")
	}

	return &config, nil

}

package config

import (
	"errors"
	"os"

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
	FilePath          string         `toml:"file_path"`
	WindowScaleFactor float64        `toml:"window_scale_factor"`
	UpdateIntervalSec float64        `toml:"update_interval_sec"`
	SpeedMultiplier   float64        `toml:"speed_multiplier"`
	OnScreenDisplay   VideoOSDConfig `toml:"OSD"`
}

type VideoOSDConfig struct {
	DisplayCycleSpeed    bool `toml:"display_cycle_speed"`
	DisplayPlaybackSpeed bool `toml:"display_playback_speed"`
	ShowOSD              bool
}

// Constants for valid configuration values
const (
	logLevelDebug = "debug"
	logLevelInfo  = "info"
	logLevelWarn  = "warn"
	logLevelError = "error"
	logLevelFatal = "fatal"

	SpeedUnitsKMH = "km/h"
	SpeedUnitsMPH = "mph"
)

// LoadFile loads the application configuration from the given filepath
func LoadFile(filepath string) (*Config, error) {

	var config Config

	// Read the TOML configuration file
	if _, err := toml.DecodeFile(filepath, &config); err != nil {
		return nil, err
	}

	// Validate all configuration elements
	if err := config.validate(); err != nil {
		return nil, err
	}

	return &config, nil

}

// validate performs validation on the configuration values
func (c *Config) validate() error {

	// Validate application configuration elements
	if err := c.App.validate(); err != nil {
		return err
	}

	// Validate remaining configuration elements
	if err := c.Speed.validate(); err != nil {
		return err
	}

	if err := c.BLE.validate(); err != nil {
		return err
	}

	if err := c.Video.validate(); err != nil {
		return err
	}

	return nil

}

// validate validates AppConfig elements
func (ac *AppConfig) validate() error {

	switch ac.LogLevel {
	case logLevelDebug, logLevelInfo, logLevelWarn, logLevelError, logLevelFatal:
		return nil
	default:
		return errors.New("invalid log level: " + ac.LogLevel)
	}

}

// validate validates SpeedConfig elements
func (sc *SpeedConfig) validate() error {

	switch sc.SpeedUnits {
	case SpeedUnitsKMH, SpeedUnitsMPH:
		return nil
	default:
		return errors.New("invalid speed units: " + sc.SpeedUnits)
	}

}

// validate validates BLEConfig elements
func (bc *BLEConfig) validate() error {

	if bc.SensorUUID == "" {
		return errors.New("sensor UUID must be specified in configuration")
	}

	return nil

}

// validate validates VideoConfig elements
func (vc *VideoConfig) validate() error {

	// Check if the video file exists
	if _, err := os.Stat(vc.FilePath); err != nil {
		return err
	}

	// Confirm that update_interval_sec is >0.0
	if vc.UpdateIntervalSec <= 0.0 {
		return errors.New("update_interval_sec must be greater than 0.0")
	}

	// Check if at least one OSD display flag is set
	vc.OnScreenDisplay.ShowOSD = (vc.OnScreenDisplay.DisplayCycleSpeed || vc.OnScreenDisplay.DisplayPlaybackSpeed)

	return nil

}

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Constants for valid configuration values
const (
	// Log levels
	logLevelDebug = "debug"
	logLevelInfo  = "info"
	logLevelWarn  = "warn"
	logLevelError = "error"
	logLevelFatal = "fatal"

	// Speed units
	SpeedUnitsKMH = "km/h"
	SpeedUnitsMPH = "mph"
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

// VideoOSDConfig represents the on-screen display configuration
type VideoOSDConfig struct {
	DisplayCycleSpeed    bool `toml:"display_cycle_speed"`
	DisplayPlaybackSpeed bool `toml:"display_playback_speed"`
	ShowOSD              bool
}

// VideoConfig represents the MPV video player configuration
type VideoConfig struct {
	FilePath          string         `toml:"file_path"`
	WindowScaleFactor float64        `toml:"window_scale_factor"`
	UpdateIntervalSec float64        `toml:"update_interval_sec"`
	SpeedMultiplier   float64        `toml:"speed_multiplier"`
	OnScreenDisplay   VideoOSDConfig `toml:"OSD"`
}

// LoadFile attempts to load the TOML configuration file from the specified path,
// falling back to the default configuration directory if not found
func LoadFile(filename string) (*Config, error) {

	// Define configuration file paths
	paths := []string{
		filename,
		filepath.Join("internal", "configuration", filepath.Base(filename)),
	}

	var lastErr error

	// Attempt to load the configuration file from each path
	for _, path := range paths {
		cfg := &Config{}

		// Load TOML file
		if _, err := toml.DecodeFile(path, cfg); err != nil {
			if !os.IsNotExist(err) || path == paths[len(paths)-1] {
				lastErr = fmt.Errorf("failed to load config from %s: %w", path, err)
			}
			continue
		}

		// Validate TOML file
		if err := cfg.validate(); err != nil {
			return nil, err
		}

		// Successfully loaded TOML file
		return cfg, nil
	}

	// Failed to load TOML file
	return nil, lastErr
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

	// Validate log level
	switch ac.LogLevel {
	case logLevelDebug, logLevelInfo, logLevelWarn, logLevelError:
		return nil
	default:
		return errors.New("invalid log level: " + ac.LogLevel)
	}
}

// validate validates BLEConfig elements
func (bc *BLEConfig) validate() error {

	// Check if the sensor UUID is specified
	if bc.SensorUUID == "" {
		return errors.New("sensor UUID must be specified in configuration")
	}

	return nil
}

// validate validates SpeedConfig elements
func (sc *SpeedConfig) validate() error {

	// Validate speed units
	switch sc.SpeedUnits {
	case SpeedUnitsKMH, SpeedUnitsMPH:
		return nil
	default:
		return errors.New("invalid speed units: " + sc.SpeedUnits)
	}
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

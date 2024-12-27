// Package config provides configuration management for the application,
// including loading and validation of TOML configuration files
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Configuration constants
const (
	// Log levels
	logLevelDebug = "debug"
	logLevelInfo  = "info"
	logLevelWarn  = "warn"
	logLevelError = "error"
	logLevelFatal = "fatal"

	// Speed units
	SpeedUnitsKMH = "km/h" // Kilometers per hour
	SpeedUnitsMPH = "mph"  // Miles per hour
)

// Config represents the complete application configuration structure
type Config struct {
	App   AppConfig   `toml:"app"`
	BLE   BLEConfig   `toml:"ble"`
	Speed SpeedConfig `toml:"speed"`
	Video VideoConfig `toml:"video"`
}

// AppConfig defines application-wide settings
type AppConfig struct {
	LogLevel string `toml:"logging_level"`
}

// BLEConfig defines Bluetooth Low Energy settings
type BLEConfig struct {
	SensorUUID      string `toml:"sensor_uuid"`
	ScanTimeoutSecs int    `toml:"scan_timeout_secs"`
}

// SpeedConfig defines speed calculation and measurement settings
type SpeedConfig struct {
	SmoothingWindow      int     `toml:"smoothing_window"`
	SpeedThreshold       float64 `toml:"speed_threshold"`
	WheelCircumferenceMM int     `toml:"wheel_circumference_mm"`
	SpeedUnits           string  `toml:"speed_units"`
}

// VideoConfig defines video playback and display settings
type VideoConfig struct {
	FilePath          string         `toml:"file_path"`
	WindowScaleFactor float64        `toml:"window_scale_factor"`
	UpdateIntervalSec float64        `toml:"update_interval_sec"`
	SpeedMultiplier   float64        `toml:"speed_multiplier"`
	OnScreenDisplay   VideoOSDConfig `toml:"OSD"`
}

// VideoOSDConfig defines on-screen display settings for video playback
type VideoOSDConfig struct {
	DisplayCycleSpeed    bool `toml:"display_cycle_speed"`
	DisplayPlaybackSpeed bool `toml:"display_playback_speed"`
	ShowOSD              bool // Computed field based on display settings
}

// LoadFile attempts to load and validate a TOML configuration file, trying multiple paths and
// returns the first valid configuration found
func LoadFile(filename string) (*Config, error) {

	paths := []string{
		filename,
		filepath.Join("internal", "configuration", filepath.Base(filename)),
	}

	var lastErr error

	for _, path := range paths {
		cfg := &Config{}

		if _, err := toml.DecodeFile(path, cfg); err != nil {

			if !os.IsNotExist(err) || path == paths[len(paths)-1] {
				lastErr = fmt.Errorf("failed to load config from %s: %w", path, err)
			}

			continue
		}

		if err := cfg.validate(); err != nil {
			return nil, err
		}

		return cfg, nil
	}

	return nil, lastErr
}

// validate performs validation across all configuration sections
func (c *Config) validate() error {

	validators := []struct {
		validate func() error
		name     string
	}{
		{c.App.validate, "app"},
		{c.Speed.validate, "speed"},
		{c.BLE.validate, "BLE"},
		{c.Video.validate, "video"},
	}

	for _, v := range validators {

		if err := v.validate(); err != nil {
			return fmt.Errorf("%s configuration error: %w", v.name, err)
		}

	}

	return nil
}

// validate checks AppConfig for valid settings
func (ac *AppConfig) validate() error {

	validLogLevels := map[string]bool{
		logLevelDebug: true,
		logLevelInfo:  true,
		logLevelWarn:  true,
		logLevelError: true,
		logLevelFatal: true,
	}

	if !validLogLevels[ac.LogLevel] {
		return fmt.Errorf("invalid log level: %v", ac.LogLevel)
	}

	return nil
}

// validate checks BLEConfig for valid settings
func (bc *BLEConfig) validate() error {

	if bc.SensorUUID == "" {
		return fmt.Errorf("sensor UUID must be specified in configuration")
	}

	return nil
}

// validate checks SpeedConfig for valid settings
func (sc *SpeedConfig) validate() error {

	validSpeedUnits := map[string]bool{
		SpeedUnitsKMH: true,
		SpeedUnitsMPH: true,
	}

	if !validSpeedUnits[sc.SpeedUnits] {
		return fmt.Errorf("invalid speed units: %v", sc.SpeedUnits)
	}

	return nil
}

// validate checks VideoConfig for valid settings
func (vc *VideoConfig) validate() error {

	if _, err := os.Stat(vc.FilePath); err != nil {
		return fmt.Errorf("video file error: %v", err)
	}

	if vc.UpdateIntervalSec <= 0.0 {
		return fmt.Errorf("update_interval_sec must be greater than 0.0")
	}

	// Set computed field based on display settings
	vc.OnScreenDisplay.ShowOSD = vc.OnScreenDisplay.DisplayCycleSpeed ||
		vc.OnScreenDisplay.DisplayPlaybackSpeed

	return nil
}

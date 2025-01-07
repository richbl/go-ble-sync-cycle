// Package config provides configuration management for the application,
// including loading and validation of TOML configuration files
package config

import (
	"flag"
	"fmt"
	"os"

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
	FontSize             int  `toml:"font_size"`
	DisplayCycleSpeed    bool `toml:"display_cycle_speed"`
	DisplayPlaybackSpeed bool `toml:"display_playback_speed"`
	DisplayTimeRemaining bool `toml:"display_time_remaining"`
	ShowOSD              bool // Computed field based on display settings
}

// LoadFile loads the configuration from a TOML file, checking first for the "-config" or "-c"
// command-line flag, falling back to the current working directory if flag not provided
func LoadFile(configFile string) (*Config, error) {

	// Parse command-line arguments
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to the configuration file")
	flag.StringVar(&configPath, "c", "", "Path to the configuration file")
	flag.Parse()

	// Use the command-line config path if provided, otherwise use the default
	if configPath != "" {
		configFile = configPath
	}

	// Read the configuration file
	cfg := &Config{}
	_, err := toml.DecodeFile(configFile, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	// Validate the configuration
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
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

	// Set ShowOSD flag based on display settings
	vc.OnScreenDisplay.ShowOSD = vc.OnScreenDisplay.DisplayCycleSpeed ||
		vc.OnScreenDisplay.DisplayPlaybackSpeed || vc.OnScreenDisplay.DisplayTimeRemaining

	return nil
}

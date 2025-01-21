package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"

	flags "github.com/richbl/go-ble-sync-cycle/internal/flags"
)

// Configuration constants
const (
	logLevelDebug = "debug"
	logLevelInfo  = "info"
	logLevelWarn  = "warn"
	logLevelError = "error"
	logLevelFatal = "fatal"

	SpeedUnitsKMH = "km/h"
	SpeedUnitsMPH = "mph"

	errFormat = "%w: %v"
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
	SeekToPosition    string         `toml:"seek_to_position"`
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

// Error messages
var (
	errInvalidLogLevel   = fmt.Errorf("invalid log level")
	errNoSensorUUID      = fmt.Errorf("sensor UUID must be specified in configuration")
	errInvalidSpeedUnits = fmt.Errorf("invalid speed units")
	errVideoFile         = fmt.Errorf("video file error")
	errInvalidInterval   = fmt.Errorf("update_interval_sec must be greater than 0.0")
	errInvalidSeek       = fmt.Errorf("seek_to_position must be in MM:SS or SS format")
)

// Load loads the configuration from a TOML file using the provided flags
func Load(configFile string) (*Config, error) {

	// Use the command-line config path if provided, otherwise use the default
	clFlags := flags.GetFlags()
	if clFlags.Config != "" {
		configFile = clFlags.Config
	}

	// Read the configuration file
	cfg := &Config{}
	_, err := toml.DecodeFile(configFile, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	// Validate configuration settings imported from the TOML file
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	// Use the command-line seek position if provided, otherwise use what's in the TOML file
	if clFlags.Seek != "" {

		if !validateTimeFormat(clFlags.Seek) {
			return nil, fmt.Errorf(errFormat, errInvalidSeek, clFlags.Seek)
		}

		cfg.Video.SeekToPosition = clFlags.Seek
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
		return fmt.Errorf(errFormat, errInvalidLogLevel, ac.LogLevel)
	}

	return nil
}

// validate checks BLEConfig for valid settings
func (bc *BLEConfig) validate() error {

	if bc.SensorUUID == "" {
		return fmt.Errorf(errFormat, errNoSensorUUID, bc.SensorUUID)
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
		return fmt.Errorf(errFormat, errInvalidSpeedUnits, sc.SpeedUnits)
	}

	return nil
}

// validate checks VideoConfig for valid settings
func (vc *VideoConfig) validate() error {

	if _, err := os.Stat(vc.FilePath); err != nil {
		return fmt.Errorf(errFormat, errVideoFile, err)
	}

	if vc.UpdateIntervalSec <= 0.0 {
		return fmt.Errorf(errFormat, errInvalidInterval, vc.UpdateIntervalSec)
	}

	if !validateTimeFormat(vc.SeekToPosition) {
		return fmt.Errorf(errFormat, errInvalidSeek, vc.SeekToPosition)
	}

	// Set ShowOSD flag based on display settings
	vc.OnScreenDisplay.ShowOSD = vc.OnScreenDisplay.DisplayCycleSpeed ||
		vc.OnScreenDisplay.DisplayPlaybackSpeed || vc.OnScreenDisplay.DisplayTimeRemaining

	return nil
}

// validateTimeFormat checks if the provided string is valid time in MM:SS or SS format
func validateTimeFormat(input string) bool {

	input = strings.TrimSpace(input)

	if strings.Contains(input, ":") {
		return validateMMSSFormat(input)
	}

	return validateSSFormat(input)
}

// validateMMSSFormat checks if the provided string is a valid time in MM:SS format
func validateMMSSFormat(input string) bool {

	// Split the input into minutes and seconds
	parts := strings.SplitN(input, ":", 2)
	if len(parts) != 2 {
		return false
	}

	minutesStr := parts[0]
	secondsStr := parts[1]

	// Validate for minutes... as long as they aren't negative
	minutes, err := strconv.Atoi(minutesStr)
	if err != nil || minutes < 0 {
		return false
	}

	// Validate for seconds (0-59)
	seconds, err := strconv.Atoi(secondsStr)
	if err != nil || seconds < 0 || seconds > 59 {
		return false
	}

	return true
}

// validateSSFormat checks if the provided string is a valid time in SS format
func validateSSFormat(input string) bool {

	// Validate for seconds... as long as they aren't negative
	seconds, err := strconv.Atoi(input)
	if err != nil || seconds < 0 {
		return false
	}

	return true
}

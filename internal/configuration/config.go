package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"

	flags "github.com/richbl/go-ble-sync-cycle/internal/flags"
)

// Config represents the complete application configuration structure from the TOML config file
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

// ValidationType, used for config validation, is a type that can be either an int or a float64
type ValidationType interface {
	int | float64
}

// validationRange is a struct used for validating config field ranges
type validationRange struct {
	value  any
	min    any
	max    any
	errMsg error
}

// Configuration constants
const (
	logLevelDebug = "debug"
	logLevelInfo  = "info"
	logLevelWarn  = "warn"
	logLevelError = "error"
	logLevelFatal = "fatal"

	SpeedUnitsKMH = "km/h"
	SpeedUnitsMPH = "mph"

	MediaPlayerMPV = "mpv"
	MediaPlayerVLC = "vlc"

	errTypeFormat = "%w: got %T"
	errFormat     = "%w: %v"
)

// Error messages
var (
	errInvalidLogLevel    = fmt.Errorf("invalid log level")
	errInvalidSpeedUnits  = fmt.Errorf("invalid speed units")
	errVideoFile          = fmt.Errorf("video file error")
	errInvalidPlayer      = fmt.Errorf("invalid media player")
	errInvalidInterval    = fmt.Errorf("update_interval_sec must be 0.1-3.0")
	errInvalidSeek        = fmt.Errorf("seek_to_position must be in MM:SS or SS format")
	errSmoothingWindow    = fmt.Errorf("smoothing window must be 1-25")
	errWheelCircumference = fmt.Errorf("wheel_circumference_mm must be 50-3000")
	errSpeedThreshold     = fmt.Errorf("speed_threshold must be 0.00-10.00")
	errSpeedMultiplier    = fmt.Errorf("speed_multiplier must be 0.1-1.5")
	errInvalidBDAddr      = fmt.Errorf("invalid sensor BD_ADDR in configuration")
	errInvalidScanTimeout = fmt.Errorf("scan_timeout_secs must be 1-100")
	errFontSize           = fmt.Errorf("font_size must be 10-200")
	errWindowScale        = fmt.Errorf("window_scale_factor must be 0.1-1.0")
	errUnsupportedType    = fmt.Errorf("unsupported type")
)

// Load loads the configuration from a TOML file using the provided flags
func Load(configFile string) (*Config, error) {
	clFlags := flags.GetFlags()

	if clFlags.Config != "" {
		configFile = clFlags.Config
	}

	cfg, err := readConfigFile(configFile, &Config{})
	if err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	if err := setSeekToPosition(cfg, clFlags); err != nil {
		return nil, err
	}

	return cfg, nil
}

// readConfigFile reads the configuration file
func readConfigFile(path string, cfg *Config) (*Config, error) {

	_, err := toml.DecodeFile(path, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return cfg, nil
}

// setSeekToPosition validates and then sets the seek position based on the command-line flag
func setSeekToPosition(cfg *Config, clFlags flags.Flags) error {

	if clFlags.Seek != "" {
		if !validateTimeFormat(clFlags.Seek) {
			return fmt.Errorf(errFormat, errInvalidSeek, clFlags.Seek)
		}
		cfg.Video.SeekToPosition = clFlags.Seek
	}

	return nil
}

// validate performs validation across all components
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

// validateConfigFields validates multiple fields against their min/max values
func validateConfigFields(validations *[]validationRange) error {

	for _, v := range *validations {
		if err := validateField(v.value, v.min, v.max, v.errMsg); err != nil {
			return err
		}
	}

	return nil
}

// validateField checks if the provided value is within the specified range
func validateField(value, minVal, maxVal any, errMsg error) error {

	switch v := value.(type) {
	case int:
		return validateInteger(v, minVal, maxVal, errMsg)
	case float64:
		return validateFloat(v, minVal, maxVal, errMsg)
	default:
		return fmt.Errorf(errTypeFormat, errUnsupportedType, v)
	}

}

// validateInteger checks if an integer value is within the specified range
func validateInteger(value int, minVal, maxVal any, errMsg error) error {

	valueMin, ok := minVal.(int)
	if !ok {
		return fmt.Errorf(errTypeFormat, errUnsupportedType, minVal)
	}

	valueMax, ok := maxVal.(int)
	if !ok {
		return fmt.Errorf(errTypeFormat, errUnsupportedType, maxVal)
	}

	return validateRange(value, valueMin, valueMax, errMsg)
}

// validateFloat checks if a float value is within the specified range
func validateFloat(value float64, minVal, maxVal any, errMsg error) error {

	valueMin, ok := minVal.(float64)
	if !ok {
		return fmt.Errorf(errTypeFormat, errUnsupportedType, minVal)
	}

	valueMax, ok := maxVal.(float64)
	if !ok {
		return fmt.Errorf(errTypeFormat, errUnsupportedType, maxVal)
	}

	return validateRange(value, valueMin, valueMax, errMsg)
}

// validateRange checks if a numeric value is within the specified range
func validateRange[T ValidationType](value, minVal, maxVal T, errMsg error) error {

	if value < minVal || value > maxVal {
		return fmt.Errorf(errFormat, errMsg, value)
	}

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

	parts := strings.SplitN(input, ":", 2)
	if len(parts) != 2 {
		return false
	}

	minutesStr := parts[0]
	secondsStr := parts[1]

	minutes, err := strconv.Atoi(minutesStr)
	if err != nil || minutes < 0 {
		return false
	}

	seconds, err := strconv.Atoi(secondsStr)
	if err != nil || seconds < 0 || seconds > 59 {
		return false
	}

	return true
}

// validateSSFormat checks if the provided string is a valid time in SS format
func validateSSFormat(input string) bool {

	seconds, err := strconv.Atoi(input)
	if err != nil || seconds < 0 {
		return false
	}

	return true
}

package config

import (
	"fmt"
	"os"
	"regexp"
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
	SensorBDAddr    string `toml:"sensor_bd_addr"`
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

// ValidationType, used for config validation, is a type that can be either an int or a float64
type ValidationType interface {
	int | float64
}

// Error messages
var (
	errInvalidLogLevel    = fmt.Errorf("invalid log level")
	errInvalidSpeedUnits  = fmt.Errorf("invalid speed units")
	errVideoFile          = fmt.Errorf("video file error")
	errInvalidInterval    = fmt.Errorf("update_interval_sec must be 0.1-3.0")
	errInvalidSeek        = fmt.Errorf("seek_to_position must be in MM:SS or SS format")
	errSmoothingWindow    = fmt.Errorf("smoothing window must be 1-25")
	errWheelCircumference = fmt.Errorf("wheel_circumference_mm must be 50-3000")
	errSpeedThreshold     = fmt.Errorf("speed_threshold must be 0.00-10.00")
	errSpeedMultiplier    = fmt.Errorf("speed_multiplier must be 0.1-1.0")
	errInvalidBDAddr      = fmt.Errorf("invalid sensor BD_ADDR in configuration")
	errInvalidScanTimeout = fmt.Errorf("scan_timeout_secs must be 1-100")
	errFontSize           = fmt.Errorf("font_size must be 10-200")
	errWindowScale        = fmt.Errorf("window_scale_factor must be 0.1-1.0")
	errUnsupportedType    = fmt.Errorf("unsupported type")
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

	// Validate the scan timeout
	if err := validateField(bc.ScanTimeoutSecs, 1, 100, errInvalidScanTimeout); err != nil {
		return err
	}

	// Define and compile the regex for BD_ADDR
	pattern := `^([0-9A-Fa-f]{2}(:[0-9A-Fa-f]{2}){5})$`
	re := regexp.MustCompile(pattern)

	// Check if the test string matches the pattern
	if !re.MatchString(strings.TrimSpace(bc.SensorBDAddr)) {
		return fmt.Errorf(errFormat, errInvalidBDAddr, bc.SensorBDAddr)
	}

	return nil
}

// validateConfigFields validates multiple fields against their min/max values
func validateConfigFields(validations []struct {
	value  any
	min    any
	max    any
	errMsg error
}) error {

	for _, v := range validations {

		if err := validateField(v.value, v.min, v.max, v.errMsg); err != nil {
			return err
		}

	}

	return nil
}

// validate checks SpeedConfig for valid settings
func (sc *SpeedConfig) validate() error {
	// Validate the speed units
	validSpeedUnits := map[string]bool{
		SpeedUnitsKMH: true,
		SpeedUnitsMPH: true,
	}

	if !validSpeedUnits[sc.SpeedUnits] {
		return fmt.Errorf(errFormat, errInvalidSpeedUnits, sc.SpeedUnits)
	}

	// Create a slice of validations to check
	validations := []struct {
		value  any
		min    any
		max    any
		errMsg error
	}{
		{sc.SmoothingWindow, 1, 25, errSmoothingWindow},
		{sc.SpeedThreshold, 0.0, 10.0, errSpeedThreshold},
		{sc.WheelCircumferenceMM, 50, 3000, errWheelCircumference},
	}

	return validateConfigFields(validations)
}

// validate checks VideoConfig for valid settings
func (vc *VideoConfig) validate() error {
	// Check if the file exists
	if _, err := os.Stat(vc.FilePath); err != nil {
		return fmt.Errorf(errFormat, errVideoFile, err)
	}

	// Create a slice of validations to check
	validations := []struct {
		value  any
		min    any
		max    any
		errMsg error
	}{
		{vc.WindowScaleFactor, 0.1, 1.0, errWindowScale},
		{vc.UpdateIntervalSec, 0.1, 3.0, errInvalidInterval},
		{vc.SpeedMultiplier, 0.1, 1.0, errSpeedMultiplier},
		{vc.OnScreenDisplay.FontSize, 10, 200, errFontSize},
	}

	if err := validateConfigFields(validations); err != nil {
		return err
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

// validateField checks if the provided value is within the specified range
func validateField(value, minVal, maxVal any, errMsg error) error {

	switch v := value.(type) {
	case int:
		return validateInteger(v, minVal, maxVal, errMsg)
	case float64:
		return validateFloat(v, minVal, maxVal, errMsg)
	default:
		return fmt.Errorf(errFormat, errUnsupportedType, v)
	}

}

// validateInteger checks if an integer value is within the specified range
func validateInteger(value int, minVal, maxVal any, errMsg error) error {

	valueMin, ok := minVal.(int)
	if !ok {
		return fmt.Errorf(errFormat, errUnsupportedType, minVal)
	}

	valueMax, ok := maxVal.(int)
	if !ok {
		return fmt.Errorf(errFormat, errUnsupportedType, maxVal)
	}

	return validateRange(value, valueMin, valueMax, errMsg)
}

// validateFloat checks if a float value is within the specified range
func validateFloat(value float64, minVal, maxVal any, errMsg error) error {

	valueMin, ok := minVal.(float64)
	if !ok {
		return fmt.Errorf(errFormat, errUnsupportedType, minVal)
	}

	valueMax, ok := maxVal.(float64)
	if !ok {
		return fmt.Errorf(errFormat, errUnsupportedType, maxVal)
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

package config

import (
	"fmt"
)

// SpeedConfig defines speed calculation and measurement settings from the TOML config file
type SpeedConfig struct {
	SpeedUnits           string  `toml:"speed_units"`
	WheelCircumferenceMM int     `toml:"wheel_circumference_mm"`
	SpeedThreshold       float64 `toml:"speed_threshold"`
	SmoothingWindow      int     `toml:"smoothing_window"`
}

// validate checks SpeedConfig for valid settings
func (sc *SpeedConfig) validate() error {

	validSpeedUnits := map[string]bool{
		SpeedUnitsKMH: true,
		SpeedUnitsMPH: true,
	}

	if !validSpeedUnits[sc.SpeedUnits] {
		return fmt.Errorf(errFormatRev, errInvalidSpeedUnits, sc.SpeedUnits)
	}

	return validateConfigFields(sc.configValidationRanges())
}

// configValidationRanges returns validation ranges for SpeedConfig
func (sc *SpeedConfig) configValidationRanges() *[]validationRange {

	return &[]validationRange{
		{sc.SmoothingWindow, 1, 25, errSmoothingWindow},
		{sc.SpeedThreshold, 0.0, 10.0, errSpeedThreshold},
		{sc.WheelCircumferenceMM, 50, 3000, errWheelCircumference},
	}
}
